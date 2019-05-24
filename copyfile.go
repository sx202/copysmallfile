package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

//全局变量定义
//sl2 可变长切片，用于记录所有文件的目录和文件名
var sl2 []string

//sl2Dir 可变长切片，用于记录所有的目录
var sl2Dir []string

//ch
//var ch chan int = make(chan int)

//错误处理
func handleError( err error)  {
	fmt.Println("Error:",err)
	os.Exit(-1)
}

//判断输入的源目录是否存在，不存在报错直接退出
func srcDirJudge(src string)  {
	_,err:=os.Stat(src)
	if err!=nil{
		handleError(err)
	}else {
		fmt.Println("源目录准备好！")
	}
}

//判断目标目录是否存在，不存在创建，存在不管
func dstDirJudge(dst string)  {
	_,err:=os.Stat(dst)
	if err!=nil{
		err = os.MkdirAll(dst,0755)
		if err!=nil{
			fmt.Println("目标目录创建失败!")
			handleError(err)
		}else {
			fmt.Println("目标目录准备好!")
		}
	}
}

//遍历目录，获取文件路径及文件名
func walkFn (path string, info os.FileInfo, err error) error{
	if info==nil{
		return err
	}
	if info.IsDir(){
		sl2Dir=append(sl2Dir,path)
		return nil
	}
	sl2 = append(sl2,path)
	return nil
}

//并发设置
func getRoutines(routines int)(sliceFileNums []int)  {


	length := len(sl2)
	fmt.Println(length)

	if length < routines{
		fmt.Println("Error:文件数比并发数少，请修改并发数！")
		os.Exit(-1)
	}

	for i:=1;i<routines;i++ {
	 	sliceFileNums = append(sliceFileNums, (length/routines)*i)
	 }
	sliceFileNums = append(sliceFileNums,length)

	return sliceFileNums
}

//创建目录：拷贝文件之前需要先把目录创建出来
func createDir(listDir []string,srcdir,dstdir string)  {
	for i:=0;i<len(listDir);i++{
		listDirTemp:=listDir[i]
		listDirTemp=strings.Replace(listDirTemp,srcdir,dstdir,-1)
		err := os.MkdirAll(listDirTemp,0755)
		if err != nil{
			fmt.Println(err)
		}
		//fmt.Println(listDir[i])
	}

}

//拷贝整个目录下所有文件
func copyDir(srcdir, dstdir string,qiepian []string,wg *sync.WaitGroup){

	//实现拷贝
	for i:=0;i<len(qiepian);i++ {

		//获取源地址
		src, err := os.Open(qiepian[i])
		if err != nil{
			handleError(err)
		}
		//defer src.Close()
		src.Close()

		//获取目标地址
		var dstname string
		dstname = strings.Replace(qiepian[i],srcdir,dstdir,1)
		dst, err := os.OpenFile(dstname,os.O_WRONLY|os.O_CREATE,0644)
		if err != nil{
			handleError(err)
		}
		//defer dst.Close()
		dst.Close()

		//拷贝文件
		io.Copy(dst,src)

		//打印是否成功
		fmt.Printf("%s 拷贝成功！\n", qiepian[i])
		//time.Sleep(time.Second)
	}
	wg.Done()
	//ch <- 0
}

//// 拷贝文件
//func copyFile(srcFile, dstFile string){
//
//	//获取源地址
//	src, err := os.Open(srcFile)
//	if err != nil{
//		handleError1(err)
//	}
//	defer src.Close()
//
//	dst, err := os.OpenFile(dstFile,os.O_WRONLY|os.O_CREATE,0644)
//	if err != nil{
//		handleError1(err)
//	}
//	defer dst.Close()
//
//	//拷贝文件
//	io.Copy(dst,src)
//	return
//
//}

func main() {

	t := time.Now()

	var srcdir,dstdir string

	fmt.Println("文件拷贝系统(现在支持的是整个目录拷贝)")
	fmt.Println("说明：")
	fmt.Println("目录书写格式")
	fmt.Println("linux系统： /server/tools")
	fmt.Println("windows系统： e:\\test")
	fmt.Println("提示：路径最后不要带'/'或'\\'")

	//源目录
	fmt.Printf("请输入要拷贝的文件目录(源目录): ")
	fmt.Scanln(&srcdir)
	//判断目录的合法性
	srcDirJudge(srcdir)

	//目标目录
	fmt.Printf("请输入拷贝到的目录(目标目录): ")
	fmt.Scanln(&dstdir)
	//判断目录的合法性
	dstDirJudge(dstdir)

	//遍历目录
	filepath.Walk(srcdir,walkFn)

	//创建所有的目标目录
	createDir(sl2Dir,srcdir,dstdir)

	//并发数
	var routines int

	gotohere:
	fmt.Printf("请输入协同完成任务的并发数: ")
	fmt.Scanln(&routines)

	//判断并发数跟文件数的关系，如果并发数大于文件数，就提示并重新输入
	if routines >= len(sl2) {
		fmt.Printf("Error:您输入的并发数超过了文件数?")
		routines=0
		time.Sleep(time.Second)
		goto gotohere
	}

	//获取当前机器CPU核数
	cpuNums:=runtime.NumCPU()
	//使用一半CPU核数运行程序
	runtime.GOMAXPROCS(cpuNums/2)

	wg := sync.WaitGroup{}

	//sl1 可变长切片  记录并发所对应的文件号[0，2] ...
	//sl3 可变长切片  记录sl2中指定位置的内容
	//sl4 可变长切片  记录sl2中指定位置的内容
	//sl5 可变长切片  记录sl2中指定位置的内容
	var sl1 []int
	var sl3 []string
	var sl4 []string
	var sl5 []string
	sl1=getRoutines(routines)
	fmt.Println(sl1)
	fmt.Println(sl2)

	//文件分片后控制对每个分片的操作*****
	for i:=0;i<routines;i++{

		switch i {
		case 0:
			switch sl1[i] {

			case 1:
				continue
			case 2:
				sl3 = sl2[1:2]
				fmt.Println(sl3)
				fmt.Println(len(sl3))
				wg.Add(1)
				go copyDir(srcdir, dstdir, sl3, &wg)
			default:
				start := 1
				end := sl1[i]
				sl4 = sl2[start:end]
				fmt.Println(sl4)
				fmt.Println(len(sl4))
				wg.Add(1)
				go copyDir(srcdir, dstdir, sl4, &wg)
			}
		case routines-1:
			start := sl1[i-1]
			sl5=sl2[start:]
			fmt.Println(sl5)
			fmt.Println(len(sl5))
			wg.Add(1)
			go copyDir(srcdir, dstdir, sl5, &wg)
		default:
			start := sl1[i-1]
			end := sl1[i]
			sl5=sl2[start:end]
			fmt.Println(sl5)
			fmt.Println(len(sl5))
			wg.Add(1)
			go copyDir(srcdir, dstdir, sl5, &wg)
		}
	}

	wg.Wait()
	//<- ch

	////拷贝文件
	//var srcFile,dstFile string
	//fmt.Printf("请输入要拷贝的文件名: ")
	//fmt.Scanln(&srcFile)
	////filepath.Walk("e:/study/test",test)
	//fmt.Printf("请输入拷贝后的文件名: ")
	//fmt.Scanln(&dstFile)
	//
	//copyFile(srcFile,dstFile)

	endTime := time.Since(t)

	fmt.Println(endTime)
	fmt.Println("end")
}

