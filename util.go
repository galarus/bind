package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
	haikunator "github.com/yelinaung/go-haikunator"
	filetype "gopkg.in/h2non/filetype.v1"
)
// Log is exported to not conflict w/ log(which gofmt was giving me troubles with when using with VSCode )
var Log = logrus.New()
var currentUser string

func init() {
	// let's log output for later grepping
	//  You could set this to any `io.Writer` such as a file
	// file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY, 0666)
	// if err == nil {
	// 	Log.Out = file
	// } else {
	// 	Log.Info("Failed to log to file, using default stderr")
	// }

	open.Run("http://localhost:8080")

	// set some common variables needed by dream()
	cmd, err := exec.Command("who").CombinedOutput()
	if err != nil {
		Log.Error("failed to know who is running the app, err: ", err)
	}
	currentUser = strings.Split(string(cmd), " ")[0]
	basePath = fmt.Sprintf("/Users/%s/Desktop/bind", currentUser)
}


// makes sure we have our working dir to place all our files
func ensureBindDirs() error {
	user, err := user.Current()
	if err != nil {
		Log.Fatal(err)
	}
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		Log.Fatal(err)
	}
	Log.Info("dir of bind is ", dir)
	Log.Info("Hello " + user.Name + " <3100LTC")
	Log.Info("====")
	Log.Info("Id: " + user.Uid)
	Log.Info("Username: " + user.Username)
	Log.Info("Home Dir: " + user.HomeDir)
	Log.Info("this is the normal user: ", currentUser)
	basePath := fmt.Sprintf("/Users/%s/Desktop/bind", currentUser)
	// Log.Info("dir: ", usr, " and expanded dir: ", exp, " and basePath to be working from is ", basePath)
	Log.Info("bind folder  basePath: ", basePath)
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		err := os.Mkdir(basePath, 0777)
		if err != nil {
			Log.Error("failed os.Mkdir", err)
		}
		Log.Info("bind FOLDER was created")
	}
	// also make dirs that live inside it
	frames := fmt.Sprintf("%s/frames", basePath)
	if _, err := os.Stat(frames); os.IsNotExist(err) {
		err := os.Mkdir(frames, 0777)
		if err != nil {
			Log.Error("failed os.Mkdir", err)
		}
		Log.Info("bind frames was created")
	}
	audio := fmt.Sprintf("%s/audio", basePath)
	if _, err := os.Stat(audio); os.IsNotExist(err) {
		err := os.Mkdir(audio, 0777)
		if err != nil {
			Log.Error("failed os.Mkdir", err)
		}
		Log.Info("bind audio was created")
	}
	videos := fmt.Sprintf("%s/videos", basePath)
	if _, err := os.Stat(videos); os.IsNotExist(err) {
		err := os.Mkdir(videos, 0777)
		if err != nil {
			Log.Error("failed os.Mkdir", err)
		}
		Log.Info("bind video was created")
	}
	logs := fmt.Sprintf("%s/logs", basePath)
	if _, err := os.Stat(logs); os.IsNotExist(err) {
		err := os.Mkdir(logs, 0777)
		if err != nil {
			Log.Error("failed os.Mkdir", err)
		}
		Log.Info("bind logs was created")
	}
	return nil
}

// saveFile will save whatever file it can
func saveFile(c *gin.Context) (string, string, error) {
	file, err := c.FormFile("file")
	if err != nil {
		Log.Info("failed to get file", err)
	}
	name := strings.Split(file.Filename, ".")[0]
	path := fmt.Sprintf("%s/frames/%s", basePath, name)
	if alreadyHave(path) {
		name = renamer(name)
		path = fmt.Sprintf("$HOME/Desktop/%s", name)
		Log.Info("\nwe renamed as: ", name)
	}
	// updateUser(name, c)

	// save the file
	savedFile := fmt.Sprintf("frames/%s/%s", name, file.Filename)
	if err := c.SaveUploadedFile(file, savedFile); err != nil {
		// c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		Log.Fatal("failed to save", err)
	}
	return "", "", errors.New("failed to save file")
}

// checkFile sort out what kind of file it is and if we support it, else error
func checkFile(c *gin.Context) (string, error) {
	reader, _, err := c.Request.FormFile("file")
	if err != nil {
		Log.Info("can't get file from form", err)
		c.String(200, "file missing from upload, please try again with a file ")
		return "", errors.New("file missing from upload")
	}
	// check if it's really a gif
	fmt.Print("checking what kind of file")
	// Log.Println(uploadFile.Filename)
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		Log.Print("not buffered")
	}
	kind, unknown := filetype.Match(buf)
	if unknown != nil {
		Log.Info("Unknown file type: %s", unknown)
		return "", errors.New("bad file type, I can't work with it: ,")
	} else if kind.Extension == "video/mp4" {
		Log.Info("it's a video!, todo: implement")
		return "video/mp4", nil
	} else if kind.Extension == "image/gif" {
		return "image/gif", nil
	} else if kind.Extension == "image/jpg" {
		return "image/jpg", nil
	} else {
		return "", errors.New("don't knwo what file this is ")
	}
}

// howManyOf returns list of .ext at a path
func howManyOf(ext string, pathS string) int {
	list := make([]string, 0, 100)
	// get all gifs in deepgif dir
	err := filepath.Walk(pathS, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ext { //like .mp4 or .mov or .png
			Log.Info(path)
			list = append(list, path)
		}
		return nil
	})
	if err != nil {
		Log.Infof("walk error [%v]\n", err)
	}
	return len(list)
}

// deepGIFFIles returns list of gifs in dir deepgif
func deepGIFFiles() []string {
	list := make([]string, 0, 100)

	// get all gifs in deepgif dir
	err := filepath.Walk("deepgif", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".mp4" {
			Log.Info(path)
			list = append(list, path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("walk error [%v]\n", err)
	}
	return list
}

func alreadyHave(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
func renamer(n string) string {
	fmt.Print("\n We had to rename the file")
	h := haikunator.New(time.Now().UTC().UnixNano())
	return fmt.Sprintf("%s%s", n, h.Haikunate())
}
