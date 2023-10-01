package main

import (
	"encoding/json"
	"log" // 导入log包
	"net/http"
	"os"
	"path/filepath"
	"fmt"
	_ "embed"
	"strings"
	"archive/zip"
	"io"
	"os/exec"
	"io/ioutil"
	//"syscall"
	//"sync"
	"flag"
)

//go:embed html/index.html
var htmlContent string

//go:embed html/stalone.html
var stalonehtmlContent string

type AddonsResponse struct {
	Addons []string `json:"addons"`
}

const tmpDir = "./tmp" // tmp子目录的路径

func main() {
	// 在启动服务器之前，清空tmp目录
	clearTmpDir()
	startBackground := flag.Bool("start-background", false, "启动参数：--start-background")
	flag.Parse()

	if !*startBackground {
		fmt.Println("错误：程序未正常启动")
		os.Exit(1)
	}

	// 定义一个路由和对应的处理函数
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/getaddonslist", getAddonslistHandler)
	http.HandleFunc("/getaddons", getAddonsHandler)
	http.HandleFunc("/unloadaddon", unloadAddonHandler)
	//http.HandleFunc("/runcmd", runCommand)

	// 启动服务器并监听指定端口
	port := 15634 // 将这里的端口改为你需要的端口号
	fmt.Printf("服务器已启动，正在监听端口 %d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("启动服务器时发生错误: %v", err)
		// 在关闭服务器时，再次清空tmp目录
		clearTmpDir()
	}

	// 在关闭服务器时，再次清空tmp目录
	clearTmpDir()
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// 将嵌入的 HTML 内容写入响应中
	fmt.Fprint(w, htmlContent)
}

func getAddonslistHandler(w http.ResponseWriter, r *http.Request) {
	addonsDir := "./addons" // addons子目录的路径

	addons, err := listAddons(addonsDir)
	if err != nil {
		log.Printf("获取Addons列表失败：%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := AddonsResponse{
		Addons: addons,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Printf("生成Addons列表响应失败：%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func listAddons(addonsDir string) ([]string, error) {
	var addons []string

	// 遍历addonsDir目录下的所有文件，找出后缀为".addons"的文件，并将其文件名添加到addons数组中
	err := filepath.Walk(addonsDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(info.Name()) == ".addons" {
			addonName := strings.TrimSuffix(info.Name(), ".addons")
			addons = append(addons, addonName)
		}
		return nil
	})

	if err != nil {
		log.Printf("遍历Addons目录失败：%v", err)
		return nil, err
	}

	return addons, nil
}

func getAddonsHandler(w http.ResponseWriter, r *http.Request) {

	addonName := r.URL.Query().Get("n")

	if addonName == "" {
		http.Error(w, "Addon name not provided", http.StatusBadRequest)
		return
	}

	// 将Addon名称与addons目录下的同名文件进行解压
	addonsDir := "./addons" // addons子目录的路径

	addonZipPath := filepath.Join(addonsDir, addonName+".addons")
	tmpAddonDir := filepath.Join(tmpDir, addonName)

	err := unzipAddon(addonZipPath, tmpAddonDir)
	if err != nil {
		log.Printf("解压Addon失败：%v", err)
		http.Error(w, "Failed to unzip the addon", http.StatusInternalServerError)
		return
	}

	// 检查是否存在 main.html 文件
	mainHTMLPath := filepath.Join(tmpAddonDir, "main.html")
	_, err = os.Stat(mainHTMLPath)
	if os.IsNotExist(err) {
		// 检查是否存在 main.bat 文件
		mainBatPath := filepath.Join(tmpAddonDir, "main.bat")
		_, err = os.Stat(mainBatPath)
		if os.IsNotExist(err) {
			// 检查是否存在 main.exe 文件
			mainExePath := filepath.Join(tmpAddonDir, "main.exe")
			_, err = os.Stat(mainExePath)
			if os.IsNotExist(err) {
				// 如果 main.html、main.bat和main.exe 都不存在，则用Win文件管理器打开该目录
				cmd := exec.Command("explorer", tmpAddonDir)
				err = cmd.Start()
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to open directory with Windows Explorer: %s", err), http.StatusInternalServerError)
					return
				}

			}else{

			// 如果存在 main.exe，则弹出cmd窗口运行 main.exe
			cmd := exec.Command("cmd", "/C", mainExePath)
			//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			err = cmd.Start()
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to run main.exe: %s", err), http.StatusInternalServerError)
				return
			}

			}
		} else {
			// 如果存在 main.bat，则弹出cmd窗口运行 main.bat
			cmd := exec.Command("cmd","/c", mainBatPath)
			//cmd.Stdout=os.Stdout
			//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			err = cmd.Start()
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to run main.bat: %s", err), http.StatusInternalServerError)
				return
			}
		}
	}else{
	// 如果 main.html 文件存在，则返回该文件内容
		mainHTMLContent, err := ioutil.ReadFile(mainHTMLPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read main.html: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(mainHTMLContent))
		return

	}

	// 无论是否运行了 main.exe 或 main.bat，都返回 "stalonehtmlContent" 的HTML响应

	// 如果存在 readme.html，则加载 readme.html 的内容
	readmeFilePath := filepath.Join(tmpAddonDir, "readme.html")
	_, err = os.Stat(readmeFilePath)
	if err == nil {
		readmeContent, err := ioutil.ReadFile(readmeFilePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read readme.html: %s", err), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(readmeContent))
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, stalonehtmlContent)
}

func unzipAddon(zipPath, targetDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		log.Printf("创建目标目录失败：%v", err)
		return err
	}

	for _, file := range reader.File {
		filePath := filepath.Join(targetDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			log.Printf("创建文件目录失败：%v", err)
			return err
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			log.Printf("创建文件失败：%v", err)
			return err
		}

		rc, err := file.Open()
		if err != nil {
			log.Printf("打开Zip文件失败：%v", err)
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			log.Printf("拷贝Zip文件内容失败：%v", err)
			return err
		}
	}

	return nil
}

func clearTmpDir() {
	// 删除tmp目录及其内容
	err := os.RemoveAll(tmpDir)
	if err != nil {
		log.Printf("清空tmp目录失败：%v", err)
		// 可以在这里打印日志或进行其他处理
	}
}

func unloadAddonHandler(w http.ResponseWriter, r *http.Request) {
	addonName := r.URL.Query().Get("n")

	if addonName == "" {
		http.Error(w, "Addon name not provided", http.StatusBadRequest)
		return
	}

	// 调用clearTmpDir函数来清空指定Addon的tmp目录
	tmpAddonDir := filepath.Join(tmpDir, addonName)
	err := clearTmpDir2(tmpAddonDir)
	if err != nil {
		log.Printf("卸载Addon %s 时出现错误：%v", addonName, err)
		http.Error(w, "Failed to unload the addon", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
}

func clearTmpDir2(dirPath string) error {
	// 删除指定目录及其内容
	err := os.RemoveAll(dirPath)
	if err != nil {
		log.Printf("清空目录失败：%v", err)
		// 可以在这里进行其他错误处理
	}
	return err
}

func runCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	addonName := r.URL.Query().Get("n")

	// Parse the JSON data from the request
	type commandData struct {
		Command string `json:"command"`
		Client  string `json:"client"`
	}

	var data commandData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	// Check if the client is "cptr_toolkits"
	if data.Client != "cptr_toolkits" {
		http.Error(w, "Unauthorized client", http.StatusUnauthorized)
		return
	}

	// Get the current working directory
	originalDir, err := os.Getwd()
	if err != nil {
		http.Error(w, "Failed to get current directory", http.StatusInternalServerError)
		return
	}

	// Change the current working directory
	if err := os.Chdir(tmpDir + "/" + addonName); err != nil {
		http.Error(w, "Failed to change directory", http.StatusInternalServerError)
		return
	}

	// Run the command
	cmd := exec.Command(data.Command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to run command: %s", err), http.StatusInternalServerError)
		return
	}

	// Change back to the original directory
	if err := os.Chdir(originalDir); err != nil {
		http.Error(w, "Failed to change back to original directory", http.StatusInternalServerError)
		return
	}

	// Return the output of the command as JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(string(output))
}