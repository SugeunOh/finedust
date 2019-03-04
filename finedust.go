package finedust

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/tebeka/selenium"
	"golang.org/x/image/font/gofont/goregular"
)

type Config struct {
	From_url                string `json:"from_url"`
	To_url                  string `json:"to_url"`
	Mode_url                string `json:"mode_url"`
	Projection_url          string `json:"projection_url"`
	Wait_seconds            int64  `json:"wait_seconds"`
	Selenium_port           int    `json:"selenium_port"`
	Selenium_window_width   int    `json:"selenium_window_width"`
	Selenium_window_height  int    `json:"selenium_window_height"`
	Selenium_webdriver_path string `json:"selenium_webdriver_path"`
	Selenium_browser        string `json:"selenium_browser"`
	Frames_path             string `json:"frames_path"`
	Time_stamp              bool   `json:"time_stamp"`
}

func (c *Config) Load(confpath string) error {
	dir, _ := filepath.Split(confpath)
	// default value
	c.Wait_seconds = 1
	c.Selenium_port = 9515
	c.Selenium_window_width = 1296 // -16  => 1280
	c.Selenium_window_height = 856 // -136 =>  720
	c.Selenium_webdriver_path = filepath.Join(dir, "tools", "chromedriver.exe")
	c.Selenium_browser = "chrome"
	c.Frames_path = filepath.Join(dir, "frames")
	c.Time_stamp = true

	//open config file
	f, err := os.Open(confpath)
	if err != nil {
		return err
	}
	defer f.Close()

	byt, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	//json
	err = json.Unmarshal(byt, c)
	if err != nil {
		return err
	}

	return nil
}

func checkFromToUrl(conf Config) error {
	pat := `^#\d\d\d\d/(0\d|10|11|12)/([012]\d|30|31)/([01]\d|20|21|22|23)00Z$`
	isMat, err := regexp.MatchString(pat, conf.From_url)
	if err != nil {
		return err
	}
	if isMat == false {
		return errors.New("from_url format error: ex) #2019/02/23/0000Z")
	}
	isMat, err = regexp.MatchString(pat, conf.To_url)
	if err != nil {
		return err
	}
	if isMat == false {
		return errors.New("to_url format error: ex) #2019/02/23/2300Z")
	}
	fmt.Println("from to url match cehk")
	return nil
}

func GenerateUrls(conf Config) ([]string, error) {
	var ret []string
	// check from -to pattern
	err := checkFromToUrl(conf)
	if err != nil {
		return nil, err
	}
	// get real time from from-to strings
	fromTim, err := time.Parse(time.RFC3339,
		fmt.Sprintf("%s-%s-%sT%s:00:00+00:00",
			conf.From_url[1:5],
			conf.From_url[6:8],
			conf.From_url[9:11],
			conf.From_url[12:14]))
	if err != nil {
		return nil, err
	}
	toTim, err := time.Parse(time.RFC3339,
		fmt.Sprintf("%s-%s-%sT%s:00:00+00:00",
			conf.To_url[1:5],
			conf.To_url[6:8],
			conf.To_url[9:11],
			conf.To_url[12:14]))
	if err != nil {
		return nil, err
	}

	// generate url
	tt := fromTim
	url := fmt.Sprintf("https://earth.nullschool.net/%s/%s/%s",
		fmt.Sprintf("#%d/%02d/%02d/%02d00Z", tt.Year(), tt.Month(), tt.Day(), tt.Hour()),
		conf.Mode_url,
		conf.Projection_url,
	)
	ret = append(ret, url)
	for {
		tt = tt.Add(time.Hour)
		if tt.After(toTim) {
			break
		}
		url = fmt.Sprintf("https://earth.nullschool.net/%s/%s/%s",
			fmt.Sprintf("#%d/%02d/%02d/%02d00Z", tt.Year(), tt.Month(), tt.Day(), tt.Hour()),
			conf.Mode_url,
			conf.Projection_url,
		)
		ret = append(ret, url)
	}

	return ret, nil
}

func WaitDownloading(wd selenium.WebDriver) (bool, error) {
	elem, err := wd.FindElement(selenium.ByCSSSelector, "#status")
	if err != nil {
		panic(err)
	}
	innerText, err := elem.Text()
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Millisecond)
	if strings.Contains(innerText, "Download") {
		return false, nil
	}
	if strings.Contains(innerText, "Rendering") {
		return false, nil
	}
	return true, nil
}

func CaptureBySelenium(conf Config) error {
	// gen urls
	urls, err := GenerateUrls(conf)
	if err != nil {
		return err
	}
	// check frames folder
	if _, err := os.Stat(conf.Frames_path); os.IsNotExist(err) {
		return errors.New("frames directory not exist (location of saved capture images)")
	}

	// check selenium driver
	if _, err := os.Stat(conf.Selenium_webdriver_path); os.IsNotExist(err) {
		return errors.New("selenium webdriver not exist")
	}

	// run chrome driver
	cmd := exec.Command(conf.Selenium_webdriver_path,
		fmt.Sprintf("--port=%d", conf.Selenium_port),
		"--url-base=wd/hub",
	)
	err = cmd.Start()
	if err != nil {
		return err
	}

	fmt.Println("wait 3 seconds...")
	time.Sleep(time.Second * 3)
	fmt.Println("start request urls")

	// connection
	caps := selenium.Capabilities{"browserName": conf.Selenium_browser}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", conf.Selenium_port))
	if err != nil {
		return (err)
	}

	// get current window handle
	nm, err := wd.CurrentWindowHandle()
	if err != nil {
		return (err)
	}

	wd.ResizeWindow(nm, conf.Selenium_window_width, conf.Selenium_window_height)

	for i, url := range urls {
		fmt.Println(i, "Request", url)
		// Get Request
		if err := wd.Get(url); err != nil {
			return (err)
		}

		if err := wd.Wait(WaitDownloading); err != nil {
			return (err)
		}

		//Wait
		time.Sleep(time.Duration(conf.Wait_seconds) * time.Second)

		// capture
		byt, err := wd.Screenshot()
		if err != nil {
			return (err)
		}
		// file name generate
		fnm := strings.Replace(strings.Replace(url[30:46], "/", "_", -1), "#", "_", -1)
		savednm := filepath.Join(conf.Frames_path, fmt.Sprintf("%s.png", fnm))

		// add time stamp text
		if conf.Time_stamp {
			byt, err = AddTimeStamp(byt, 30, 30, 24, image.White, fnm)
			if err != nil {
				return err
			}
		}
		//save file
		f, err := os.Create(savednm)
		if err != nil {
			return (err)
		}
		defer f.Close()
		_, err = f.Write(byt)
		if err != nil {
			return (err)
		}

	}

	err = wd.Close()
	if err != nil {
		fmt.Println(err)
	}
	err = wd.Quit()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("wait connection closed...")
	time.Sleep(time.Second * 1)

	err = cmd.Process.Kill()
	if err != nil {
		return err
	}

	return nil
}

// input byt: png bytes =>  output byt: png bytes
func AddTimeStamp(byt []byte, x, y, size int, color color.Color, label string) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(byt))
	if err != nil {
		return nil, err
	}
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}
	face := truetype.NewFace(font, &truetype.Options{Size: float64(size)})

	w := img.Bounds().Size().X
	h := img.Bounds().Size().Y

	dc := gg.NewContext(w, h)
	dc.DrawImage(img, 0, 0)
	dc.SetFontFace(face)
	r, g, b, _ := color.RGBA()
	dc.SetRGB(
		float64(r)/float64(0xff),
		float64(g)/float64(0xff),
		float64(b)/float64(0xff),
	)
	dc.DrawStringAnchored(label, float64(x), float64(y), 0, 0.5)

	img = dc.Image()

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)

	return buf.Bytes(), nil
}

type RenameInfo struct {
	Start string   `json:"start"`
	End   string   `json:"end"`
	Files []string `json:"files"`
}

func RenameImageSequence(path, start, end string) error {
	// set start ~ end range
	if start <= "0000_00_00_0000Z.png" || start > "9999_99_99_9999Z.png" {
		start = "0000_00_00_0000Z.png"
	}
	if end > "9999_99_99_9999Z.png" || end <= "0000_00_00_0000Z.png" {
		end = "9999_99_99_9999Z.png"
	}
	// walk path
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	var strs []string

	for _, file := range files {
		strs = append(strs, file.Name())
		if strings.HasPrefix(file.Name(), "img") {
			return errors.New("Need to Rollback ImageSequence names")
		}
	}
	sort.Strings(strs)

	start_idx := 0
	end_idx := 0

	for i, v := range strs {
		if v < start {
			start_idx = i
		}
		if v < end {
			end_idx = i
		}
	}
	start = strs[start_idx]
	end = strs[end_idx]
	renameInfo := &RenameInfo{start, end, strs[start_idx:(end_idx + 1)]}

	b, err := json.Marshal(renameInfo)
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(path, "renameinfo.json"))
	if err != nil {
		return (err)
	}
	_, err = f.Write(b)
	if err != nil {
		return (err)
	}
	f.Close()

	// change names
	for i, v := range renameInfo.Files {
		err = os.Rename(
			filepath.Join(path, v),
			filepath.Join(path, fmt.Sprintf("img%04d.png", i)),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func RollbackImageSequence(path string) error {
	// check frames folder
	if _, err := os.Stat(filepath.Join(path, "renameinfo.json")); os.IsNotExist(err) {
		return nil
	}

	//open config file
	f, err := os.Open(filepath.Join(path, "renameinfo.json"))
	if err != nil {
		return err
	}

	byt, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	renameInfo := &RenameInfo{}

	//json
	err = json.Unmarshal(byt, &renameInfo)
	if err != nil {
		return err
	}

	// json file close
	f.Close()

	// change names
	for i, v := range renameInfo.Files {
		src := filepath.Join(path, fmt.Sprintf("img%04d.png", i))
		dst := filepath.Join(path, v)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}
		err = os.Rename(
			src,
			dst,
		)
		if err != nil {
			return err
		}
	}

	// delete json
	err = os.Remove(filepath.Join(path, "renameinfo.json"))
	if err != nil {
		return err
	}

	return nil
}

func GenVideoByFFMPEG(binpath, imgpath string, quality, scale int) error {
	cmd := exec.Command(
		binpath,
		"-f", "image2",
		"-r", "24",
		"-i", filepath.Join(imgpath, "img%04d.png"),
		"-q:v", fmt.Sprintf("%d", quality), //  1 (lossless), 4 (quard
		"-vf", fmt.Sprintf("scale=-1:%d", scale),
		"out.avi",
	)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	fmt.Println("start")
	err := cmd.Run()
	if err != nil {
		return err
	}
	fmt.Println("done")
	return nil
}
