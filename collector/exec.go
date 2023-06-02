package collector

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	order  int
	awkMap = make(map[int]string)
	result = make(map[string]string)
	// 定义要在status文件里筛选的关键字
	targetList   = []string{"Name", "State", "PPid", "Uid", "Gid", "VmHWM", "VmRSS"}
	targetResult = make(map[string]map[string]string)
)

func stringGrep(s string, d string) (bool, error) {
	for k, v := range d {
		if v != rune(s[k]) {
			return false, fmt.Errorf("string does not match")
		}
	}
	order = 1
	resolv, err := stringAWK(s[len(d):])
	if len(resolv) == 0 {
		return false, err
	}
	order = 0
	return true, nil
}

func stringAWK(s string) (map[int]string, error) {
	i := 0
	for k, v := range s {
		if v != rune(9) && v != rune(32) && v != rune(10) {
			i = 1
			awkMap[order] += string(v)
		} else {
			if i > 0 {
				order++
				i = 0
			}
			stringAWK(s[k+1:])
			return awkMap, nil
		}
	}
	return awkMap, fmt.Errorf("awk error")
}

func GetProcessInfo(p []string, m string) map[string]map[string]string {
	for _, port := range p {
		// 通过端口号获取进程pid信息
		// 通过组合命令行的方式执行linux命令，筛选出pid
		cmd := "sudo " + m + " -tnlp" + "|grep :" + port + "|awk '{print $NF}'|awk -F'/' '{print $1}'"
		getPid := exec.Command("bash", "-c", cmd)
		out, err := getPid.Output()
		if err != nil {
			fmt.Println("exec command failed", err)
			return nil
		}
		dir := strings.ReplaceAll(string(out), "\n", "")
		if len(dir) == 0 {
			fmt.Println("'dir' string is empty")
			return nil
			// panic("'dir' string is empty")
		}
		// fmt.Println("test_dir", dir)
		result["pid"] = dir
		// 获取命令行绝地路径
		cmdRoot := "sudo ls -l /proc/" + dir + "/exe |awk '{print $NF}'"
		getCmdRoot := exec.Command("bash", "-c", cmdRoot)
		out, err = getCmdRoot.Output()
		if err != nil {
			fmt.Println("exec getCmdRoot command failed", err)
		}
		// fmt.Println("test_cmdroot", strings.ReplaceAll(string(out), "\n", ""))
		result["cmdroot"] = strings.ReplaceAll(string(out), "\n", "")
		// 获取/proc/pid/cmdline文件内信息
		cmdline, err := os.Open("/proc/" + dir + "/cmdline")
		if err != nil {
			fmt.Println("open cmdline file error :", err)
			panic(err)
		}
		cmdlineReader, err := bufio.NewReader(cmdline).ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Println(err)
		}
		result["cmdline"] = strings.ReplaceAll(cmdlineReader, "\x00", " ")
		// 获取/proc/pid/status文件内信息
		status, err := os.Open("/proc/" + dir + "/status")
		if err != nil {
			fmt.Println("open status file error :", err)
		}

		// 执行函数返回前关闭打开的文件
		defer cmdline.Close()
		defer status.Close()

		statusReader := bufio.NewReader(status)
		if err != nil {
			fmt.Println(err)
		}

		for {
			line, err := statusReader.ReadString('\n') //注意是字符
			if err == io.EOF {
				if len(line) != 0 {
					fmt.Println(line)
				}
				break
			}
			if err != nil {
				fmt.Println("read file failed, err:", err)
				// return
			}
			for _, v := range targetList {
				istrue, _ := stringGrep(line, v)
				if istrue {
					result[v] = awkMap[2]
					// fmt.Printf("%v结果是：%v\n", v, awkMap[2])
					awkMap = make(map[int]string)
				}
			}
		}
		// fmt.Println("数据的和:", result)
		// fmt.Println("test_result", result)
		targetResult[port] = result
		// 给result map重新赋值，要不然使用的是同一个map指针，targetResult结果是一样的
		result = make(map[string]string)
	}
	// fmt.Println("test_total", targetResult)
	return targetResult
}
