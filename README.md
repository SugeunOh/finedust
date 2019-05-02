# finedust
미세먼지 영상 제작 프로그램 소스
[Earth.nullschool](https://earth.nullschool.net/) 에서 타임랩스
영상제작을 위해 연속 캡쳐사진을 찍는 매크로 프로그램

## 필요한 Tools
* [Chrome Driver](http://chromedriver.chromium.org/downloads)
* [ffmpeg](http://ffmpeg.org/download.html)

## 사용환경
* Windows 7
* Go 1.11

## 빌드
cmd/capture와 cmd/video  각각을 빌드하여 실행파일로 만든다.

## 디렉토리구조

```
\frames
\tools
    \chromedriver.exe
    \ffmpeg.exe
capture.exe
video.exe
config.json
```

위 크롬드라이버와 ffmpeg 툴을 tools 디렉토리에 넣는다.




## Capture config설정

config.json을 수정하여 사용한다.
설정한뒤 capture.exe를 실행하면 수행된다.

```json
{
	"from_url" : "#2019/04/01/0000Z",
	"to_url" : "#2019/04/30/2300Z",
	"mode_url" : "chem/surface/level/anim=off/overlay=cosc",
	"projection_url" : "orthographic=-234.04,36.14,1377",
	"wait_seconds": 1,
	"selenium_port": 9515,
	"selenium_window_width": 1296,
	"selenium_window_height": 856,
	"selenium_webdriver_path": "tools/chromedriver.exe",
	"selenium_browser" : "chrome",
	"frames_path" : "framescosc",
	"time_stamp": true
}
``` 

* from - to는 시작 시간 및 종료시간으로 한다.
* mode_url은 url중 중간 mode 와 관련된 부분이다. 현재 cosc로 설정되어있음,
particulates/surface/level/anim=off/overlay=pm10 등도 사용가능
* projection_url은 url 중 뒷부분, 위치와 보는방식에 관련된 부분이다.
* wait_seconds 는 캡쳐시 최소 인터벌이다 (인간적으로 설정할것 서버에 부하주지맙시다)
* selenium 과 관련된 부분 width와 height 는 크롬 전체 크기를 의미하며
내부 웹 페이지의 크기는 더 작으므로 시행착오를 통해 정확한 내부크기를 추정하세요
* frames_path 는 캡쳐사진을 저장할 폴더 (상대경로)
* time_stamp는 캡쳐사진에 시간을 출력할것인가를 결정


## video 사용

video -h를 치면 사용방법이 나온다.
```
bin\finedust>video -h
Usage of video:
  -binpath string
        ffmpeg path (default "tools/ffmpeg")
  -end string
        end image file
  -imgpath string
        image sequence directory (default "frames")
  -quality int
        1:lossless (default 4)
  -scale int
        height scale (default 720)
  -start string
        start image file

```

따라서 아래 예처럼 사용

```
bin\finedust>video -start=2019_04_01_0000Z.png -end=2019_04_30_2300Z.png -imgpath=framescosc
```

윈도우 ffmpeg은 glob 파일 패턴방식이 적용되지 않으므로
video 프로그램은 시작이미지부터 종료이미지까지 일괄적으로 숫자패턴의 파일명으로 
변경후 ffmpeg으로 영상을 만든뒤 원래 이름으로 복구하는 과정을 거친다.

