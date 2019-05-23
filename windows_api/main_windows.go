package main

import (
	"encoding/hex"
	"fmt"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/pkg/errors"
)

func main() {
	var order uint32 = 1 // True
	var ulAf uint32 = AF_INET
	var tableClass uint32 = 0

	buffer, err := GetExtendedTcpTable(order, ulAf, tableClass)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(hex.Dump(buffer))

	pTable := (*MIB_TCPTABLE_OWNER_PID)(unsafe.Pointer(&buffer[0]))

	rows := make([]MIB_TCPROW_OWNER_PID, int(pTable.dwNumEntries))
	for i := 0; i < int(pTable.dwNumEntries); i++ {
		rows[i] = *(*MIB_TCPROW_OWNER_PID)(unsafe.Pointer(
			uintptr(unsafe.Pointer(&pTable.table[0])) +
				uintptr(i)*unsafe.Sizeof(pTable.table[0])))
	}
	show(rows)

	rows2 := ((*[1 << 30]MIB_TCPROW_OWNER_PID)(unsafe.Pointer(&pTable.table[0]))[:int(pTable.dwNumEntries):int(pTable.dwNumEntries)])
	show(rows2)

}

var (
	kernel32DLL          = syscall.NewLazyDLL("kernel32.dll")
	procCreateJobObjectW = kernel32DLL.NewProc("CreateJobObjectW")

	iphlpapiDLL             = syscall.NewLazyDLL("iphlpapi.dll")
	procGetExtendedTcpTable = iphlpapiDLL.NewProc("GetExtendedTcpTable")
)

const (
	AF_INET = 2
)

type MIB_TCPTABLE_OWNER_PID struct {
	dwNumEntries uint32
	table        [1]MIB_TCPROW_OWNER_PID
}

type MIB_TCPROW_OWNER_PID struct {
	dwState      uint32
	dwLocalAddr  [4]byte
	dwLocalPort  uint32
	dwRemoteAddr [4]byte
	dwRemotePort uint32
	dwOwningPid  uint32
}

func show(raws []MIB_TCPROW_OWNER_PID) {
	for _, v := range raws {
		v.show()
	}
}

func (m *MIB_TCPROW_OWNER_PID) show() {
	fmt.Printf(`
	state: %d
	loadAddr: %s
	localPort: %d
	remoteAddr: %s
	remotePort: %d
	pid: %d`,
		m.dwState,
		string(m.dwLocalAddr[:]),
		m.dwLocalPort,
		string(m.dwRemoteAddr[:]),
		m.dwRemotePort,
		m.dwOwningPid)
}

// GetExtendedTcpTable function retrieves a table that contains a list of TCP endpoints available to the application.
func GetExtendedTcpTable(order, ulAf, tableClass uint32) ([]byte, error) {

	var dwSize uint32
	ret, _, err := procGetExtendedTcpTable.Call(
		0,                                // PVOID
		uintptr(unsafe.Pointer(&dwSize)), // PDWORD
		uintptr(order),                   // BOOL
		uintptr(ulAf),                    // ULONG
		uintptr(tableClass),              // TCP_TABLE_CLASS
		0,                                // ULONG
	)
	if ret == 0 {
		return nil, errors.Wrapf(err, "get extended tcp table size failed code %x", ret)
	}

	if syscall.Errno(ret) == syscall.ERROR_INSUFFICIENT_BUFFER {
		buffer := make([]byte, int(dwSize))

		ret, _, err := procGetExtendedTcpTable.Call(
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(unsafe.Pointer(&dwSize)),
			uintptr(order),
			uintptr(ulAf),
			uintptr(tableClass),
			uintptr(uint32(0)),
		)

		if ret != 0 {
			return nil, errors.Wrapf(err, "get extended tcp table failed code %x", ret)
		}

		return buffer, nil
	}

	return nil, errors.Wrapf(err, "get extended tcp table size failed code %x", ret)
}

// CreateJobObject uses the CreateJobObjectW Windows API Call to create and return a Handle to a new JobObject
func CreateJobObject(attr *syscall.SecurityAttributes, name string) (syscall.Handle, error) {
	r1, _, err := procCreateJobObjectW.Call(
		uintptr(unsafe.Pointer(attr)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name))),
	)
	if err != syscall.Errno(0) {
		return 0, err
	}
	return syscall.Handle(r1), nil
}

// StringToCharPtr converts a Go string into pointer to a null-terminated cstring.
// This assumes the go string is already ANSI encoded.
func StringToCharPtr(str string) *uint8 {
	chars := append([]byte(str), 0) // null terminated
	return &chars[0]
}

// StringToUTF16Ptr converts a Go string into a pointer to a null-terminated UTF-16 wide string.
// This assumes str is of a UTF-8 compatible encoding so that it can be re-encoded as UTF-16.
func StringToUTF16Ptr(str string) *uint16 {
	wchars := utf16.Encode([]rune(str + "\x00"))
	return &wchars[0]
}
