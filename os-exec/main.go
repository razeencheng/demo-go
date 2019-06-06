package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

func main() {

	_, err := exec.LookPath("./testcmd/testcmd")
	if err != nil {
		panic("可执行文件不可用 " + err.Error())
	}

	fmt.Println("\nRun Test 1")
	test1()

	fmt.Println("\nRun Test 2")
	test2()

	fmt.Println("\nRun Test 3")
	test3()

	fmt.Println("\nRun Test 4")
	test4()

	fmt.Println("\nRun Test 5")
	test5()
}

// 持续输入
func test5() {
	cmd := exec.Command("openssl")

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	stdin, _ := cmd.StdinPipe()

	cmd.Start()

	// 读
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := stderr.Read(buf)

			if n > 0 {
				fmt.Println(string(buf[:n]))
			}

			if n == 0 {
				break
			}

			if err != nil {
				log.Printf("read err %v", err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := stdout.Read(buf)

			if n == 0 {
				break
			}

			if n > 0 {
				fmt.Println(string(buf[:n]))
			}

			if n == 0 {
				break
			}

			if err != nil {
				log.Printf("read out %v", err)
				return
			}

		}
	}()

	// 写
	go func() {
		stdin.Write([]byte("version\n\n"))
		stdin.Write([]byte("ciphers -v\n\n"))
		stdin.Write([]byte("s_client -connect razeencheng.com:443"))
		stdin.Close()
		wg.Done()
	}()

	wg.Wait()
	err := cmd.Wait()
	if err != nil {
		log.Printf("cmd wait %v", err)
		return
	}

}

// 通过上下文控制超时
func test4() {

	ctx, calcel := context.WithTimeout(context.Background(), 2*time.Second)
	defer calcel()

	cmd := exec.CommandContext(ctx, "./testcmd/testcmd", "-s", "-e")

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	oReader := bufio.NewReader(stdout)
	eReader := bufio.NewReader(stderr)

	cmd.Start()

	go func() {
		for {
			line, err := oReader.ReadString('\n')

			if line != "" {
				log.Printf("read line %s", line)
			}

			if err != nil || line == "" {
				log.Printf("read line err %v", err)
				return
			}

		}
	}()

	go func() {
		for {
			line, err := eReader.ReadString('\n')

			if line != "" {
				log.Printf("read err %s", line)
			}

			if err != nil || line == "" {
				log.Printf("read err %v", err)
				return
			}

		}
	}()

	err := cmd.Wait()
	if err != nil {
		log.Printf("cmd wait %v", err)
		return
	}
}

// 按行读输出的内容
func test3() {
	cmd := exec.Command("./testcmd/testcmd", "-s", "-e")
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	oReader := bufio.NewReader(stdout)
	eReader := bufio.NewReader(stderr)

	cmd.Start()

	go func() {
		for {
			line, err := oReader.ReadString('\n')

			if line != "" {
				log.Printf("read line %s", line)
			}

			if err != nil || line == "" {
				log.Printf("read line err %v", err)
				return
			}

		}
	}()

	go func() {
		for {
			line, err := eReader.ReadString('\n')

			if line != "" {
				log.Printf("read err %s", line)
			}

			if err != nil || line == "" {
				log.Printf("read err %v", err)
				return
			}

		}
	}()

	err := cmd.Wait()
	if err != nil {
		log.Printf("cmd wait %v", err)
		return
	}

}

// 	stdout & stderr 分开输出
func test2() {
	cmd := exec.Command("./testcmd/testcmd", "-s", "-e")
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()

	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := stderr.Read(buf)

			if n > 0 {
				log.Printf("read err %s", string(buf[:n]))
			}

			if n == 0 {
				break
			}

			if err != nil {
				log.Printf("read err %v", err)
				return
			}
		}
	}()

	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := stdout.Read(buf)

			if n == 0 {
				break
			}

			if n > 0 {
				log.Printf("read out %s", string(buf[:n]))

			}

			if n == 0 {
				break
			}

			if err != nil {
				log.Printf("read out %v", err)
				return
			}

		}
	}()

	err := cmd.Wait()
	if err != nil {
		log.Printf("cmd wait %v", err)
		return
	}

}

// 简单执行
func test1() {
	cmd := exec.Command("./testcmd/testcmd", "-s")

	// 使用CombinedOutput 将stdout stderr合并输出
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("test1 failed %s\n", err)
	}
	log.Println("test1 output ", string(out))
}
