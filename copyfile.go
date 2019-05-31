package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

//代码写作规范
//所有的变量、方法都使用英文命名，必须要词意对应
//所有的变量名书写格式使用小驼峰
//所有的方法名书写格式使用大驼峰

//全局变量定义
//sl2 可变长切片，用于记录所有文件的目录和文件名
//sl2Dir 可变长切片，用于记录所有的目录
var sl2,sl2Dir []string


//ch
//var ch chan int = make(chan int)


//错误处理
func HandleError( err error)  {
	fmt.Println("Error:",err)
	os.Exit(-1)
}

//判断程序运行在什么系统下，根据系统给出相应的提示
func SysType()  {

	fmt.Println("文件拷贝系统（主要解决的是大量小文件的拷贝）")
	fmt.Println("说明：")

	TYPE := runtime.GOOS

	if TYPE == "linux" {
		fmt.Println("linux系统路径示例： /server/tools/")
		fmt.Println("提示：路径最后一定要带'/'")
	}

	if TYPE == "windows"{
		fmt.Println("windows系统路径示例： e:\\test\\")
		fmt.Println("提示：路径最后一定要带'\\'")
	}
}

//判断输入的源文件或目录是否存在。
// 1、不存在报错直接退出。
// 2、存在判断一下输入的是目录还是文件。
//    num = 1 代表是目录  num = 0 代表是文件
func Src_File_Dir_Judge(src string) (num int64) {

	f,err := os.Stat(src)

	//判断文件或者目录是否存在，不存在报错
	if err != nil{
		HandleError(err)
	}
	if f.IsDir() {
		fmt.Println("源目录准备好！")
		num=1
	}else {
		fmt.Println("源文件准备好！")
		num=0
	}
	return num
}

//判断目标目录是否存在，不存在创建，存在不管
func Dst_Dir_Judge(dst string)  {

	_,err := os.Stat(dst)
	if err != nil {
		err = os.MkdirAll(dst,0755)
		if err != nil {
			fmt.Println("目标目录创建失败!")
			HandleError(err)
		}else {
			fmt.Println("目标目录准备好!")
		}
	}else {
		fmt.Println("目标目录准备好!")
	}
}

//遍历目录，获取文件路径及文件名
func WalkFn (path string, info os.FileInfo, err error) error{
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
func GetRoutines(routines int)(sliceFileNums []int)  {

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
func CreateDir(listDir []string,srcdir,dstdir string)  {
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
func Copy_Dir(srcdir, dstdir string,qiepian []string,wg *sync.WaitGroup){

	//实现拷贝
	for i:=0;i<len(qiepian);i++ {

		//获取源地址
		src, err := os.Open(qiepian[i])
		if err != nil{
			HandleError(err)
		}
		//defer src.Close()
		src.Close()

		//获取目标地址
		var dstname string
		dstname = strings.Replace(qiepian[i],srcdir,dstdir,1)
		dst, err := os.OpenFile(dstname,os.O_WRONLY|os.O_CREATE,0644)
		if err != nil{
			HandleError(err)
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

//拷贝文件
func Copy_File(srcFileName,dstDirName string)  {

	//打开源文件
	srcFile,err := os.Open(srcFileName)
	if err != nil {
		HandleError(err)
	}
	defer srcFile.Close()

	//获取源文件大小
	srcFileInfo,err := os.Stat(srcFileName)
	if err != nil {
		HandleError(err)
	}
	srcFileSize := srcFileInfo.Size()

	//拼接目标文件目录+文件名
	var srcDirName string
	TYPE := runtime.GOOS
	if TYPE == "linux" {
		reg := regexp.MustCompile(`^.*/`)
		srcDirName = reg.FindString(srcFileName)
	}
	if TYPE == "windows"{
		reg := regexp.MustCompile(`^.*\\`)
		srcDirName = reg.FindString(srcFileName)
	}

	//判断文件是否存在，存在提示退出，不存在创建
	DstFileName := strings.Replace(srcFileName,srcDirName,dstDirName,-1)
	_,err = os.Stat(DstFileName)

	if err == nil {
		fmt.Println("文件已经存在，请重新选择！")
	}else {
		//创建目标文件
		DstFile, err := os.Create(DstFileName)
		if err != nil {
			HandleError(err)
		}
		defer DstFile.Close()

		//拷贝文件
		dstFileSize, err := io.Copy(DstFile, srcFile)
		if err != nil{
			HandleError(err)
		}

		//打印 源文件、拷贝后文件大小
		fmt.Printf("源文件大小：%s  拷贝后文件的大小：%s \n", srcFileSize, dstFileSize)
	}
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

	var srcDirName,dstDirName string
	var dir_File_Nums,routines,copyType int64

	//判断程序运行在什么系统下，根据系统给出相应的提示
	SysType()

	//输入源文件或目录
	fmt.Printf("请输入需要拷贝的文件或目录名称: ")
	fmt.Scanln(&srcDirName)

	//判断输入的源文件或目录是否存在。
	// 1、不存在报错直接退出。
	// 2、存在判断一下输入的是目录还是文件。
	//    返回值 1 代表是目录   0 代表是文件
	dir_File_Nums = Src_File_Dir_Judge(srcDirName)

	//目标目录
	fmt.Printf("请输入拷贝到的目录(目标目录): ")
	fmt.Scanln(&dstDirName)

	//判断目标目录是否存在，不存在创建，存在不管
	Dst_Dir_Judge(dstDirName)

	//拷贝方式选择
	fmt.Printf("本系统支持三种拷贝文件的方式：1.普通拷贝 2.加速拷贝 3.极速拷贝 ")
	fmt.Printf("请输入拷贝文件的方式（只要输入对应的数字即可）: ")
	fmt.Scanln(&copyType)

	Copy_File(srcDirName,dstDirName)

	//遍历目录
	filepath.Walk(srcDirName,WalkFn)

	//创建所有的目标目录
	CreateDir(sl2Dir,srcDirName,dstDirName)


	//gotohere:
	//fmt.Printf("请输入协同完成任务的并发数: ")
	//fmt.Scanln(&routines)
	//
	////判断并发数跟文件数的关系，如果并发数大于文件数，就提示并重新输入
	//if routines >= len(sl2) {
	//	//获取当前机器CPU核数
	//	fmt.Printf("Error:您输入的并发数超过了文件数?")
	//	routines=0
	//	time.Sleep(time.Second)
	//	goto gotohere
	//}

	cpuNums:=runtime.NumCPU()
	//使用一半CPU核数运行程序
	runtime.GOMAXPROCS(cpuNums/2)

	wg := sync.WaitGroup{}

	//sl1 可变长切片  记录并发所对应的文件号[0，2] ...
	//sl3 可变长切片  记录sl2中指定位置的内容
	//sl4 可变长切片  记录sl2中指定位置的内容
	//sl5 可变长切片  记录sl2中指定位置的内容
	//var sl1 []int
	//var sl3 []string
	//var sl4 []string
	//var sl5 []string
	//sl1=getRoutines(routines)
	//fmt.Println(sl1)
	//fmt.Println(sl2)

	////文件分片后控制对每个分片的操作*****
	//for i:=0;i<routines;i++{
	//
	//	switch i {
	//	case 0:
	//		switch sl1[i] {
	//
	//		case 1:
	//			continue
	//		case 2:
	//			sl3 = sl2[1:2]
	//			fmt.Println(sl3)
	//			fmt.Println(len(sl3))
	//			wg.Add(1)
	//			go copyDir(srcdir, dstdir, sl3, &wg)
	//		default:
	//			start := 1
	//			end := sl1[i]
	//			sl4 = sl2[start:end]
	//			fmt.Println(sl4)
	//			fmt.Println(len(sl4))
	//			wg.Add(1)
	//			go copyDir(srcdir, dstdir, sl4, &wg)
	//		}
	//	case routines-1:
	//		start := sl1[i-1]
	//		sl5=sl2[start:]
	//		fmt.Println(sl5)
	//		fmt.Println(len(sl5))
	//		wg.Add(1)
	//		go copyDir(srcdir, dstdir, sl5, &wg)
	//	default:
	//		start := sl1[i-1]
	//		end := sl1[i]
	//		sl5=sl2[start:end]
	//		fmt.Println(sl5)
	//		fmt.Println(len(sl5))
	//		wg.Add(1)
	//		go copyDir(srcdir, dstdir, sl5, &wg)
	//	}
	//}

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

