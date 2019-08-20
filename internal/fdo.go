package internal // import "fyne.io/desktop/internal"

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/desktop"
	"fyne.io/fyne"
)

var iconExtensions = []string{".png", ".svg"}

//fdoIconData is a structure that contains information about .desktop files
type fdoIconData struct {
	name     string // Application name
	iconName string // Icon name
	iconPath string // Icon path
	exec     string // Command to execute application
}

//Name returns the name associated with an fdo app
func (data *fdoIconData) Name() string {
	return data.name
}

//IconName returns the name of the icon that an fdo app wishes to use
func (data *fdoIconData) IconName() string {
	return data.iconName
}

//IconPath returns the path of the icon that an fdo app wishes to use
func (data *fdoIconData) IconPath() string {
	return data.iconPath
}

//Exec returns the command used to execute the fdo app
func (data *fdoIconData) Exec() string {
	return data.exec
}

//fdoLookupXdgDataDirs returns a string slice of all XDG_DATA_DIRS
func fdoLookupXdgDataDirs() []string {
	dataLocation := os.Getenv("XDG_DATA_DIRS")
	locationLookup := strings.Split(dataLocation, ":")
	if len(locationLookup) == 0 {
		var fallbackLocations []string
		fallbackLocations = append(fallbackLocations, "/usr/local/share")
		fallbackLocations = append(fallbackLocations, "/usr/share")
		return fallbackLocations
	}
	return locationLookup
}

//fdoLookupApplicationByMetadata looks up an application by comparing the requested name to the contents of .desktop files
func fdoLookupApplicationByMetadata(theme string, size int, appName string) desktop.IconData {
	locationLookup := fdoLookupXdgDataDirs()
	for _, dataDir := range locationLookup {
		testLocation := filepath.Join(dataDir, "applications")
		files, err := ioutil.ReadDir(testLocation)
		if err != nil {
			continue
		}
		for _, f := range files {
			if strings.HasPrefix(f.Name(), ".") || f.IsDir() {
				continue
			}
			icon := newFdoIconData(theme, size, filepath.Join(testLocation, f.Name()))
			if icon == nil {
				return icon
			}
			if icon.name == appName || icon.exec == appName {
				return icon
			}
		}
	}
	return nil
}

//fdoLookupApplication looks up an application by name and returns an fdoIconData struct
func fdoLookupApplication(theme string, size int, appName string) desktop.IconData {
	locationLookup := fdoLookupXdgDataDirs()
	for _, dataDir := range locationLookup {
		testLocation := filepath.Join(dataDir, "applications", appName+".desktop")
		if _, err := os.Stat(testLocation); err == nil {
			return newFdoIconData(theme, size, testLocation)
		}
	}
	//If no match was found checking by filenames, check by file contents
	return fdoLookupApplicationByMetadata(theme, size, appName)
}

//fdoLookupApplicationPartial looks up an application by a partial name and returns all matches in an fdoIconData struct slice
func fdoLookupApplicationMatching(theme string, size int, appName string) []desktop.IconData {
	var icons []desktop.IconData
	locationLookup := fdoLookupXdgDataDirs()
	for _, dataDir := range locationLookup {
		testLocation := filepath.Join(dataDir, "applications")
		files, err := ioutil.ReadDir(testLocation)
		if err != nil {
			continue
		}
		for _, f := range files {
			if strings.HasPrefix(f.Name(), ".") || f.IsDir() {
				continue
			}
			icon := newFdoIconData(theme, size, filepath.Join(testLocation, f.Name()))
			if icon == nil {
				continue
			}
			if strings.Contains(strings.ToLower(icon.name), strings.ToLower(appName)) ||
				strings.Contains(strings.ToLower(icon.exec), strings.ToLower(appName)) {
				icons = append(icons, icon)
			}
		}
	}
	return icons
}

//fdoLookupApplicationWinInfo looks up an application based on window info and returns an fdoIconData struct
func fdoLookupApplicationWinInfo(theme string, size int, win desktop.Window) desktop.IconData {
	icon := fdoLookupApplication(theme, size, win.Title())
	if icon != nil {
		return icon
	}
	for _, class := range win.Class() {
		icon := fdoLookupApplication(theme, size, class)
		if icon != nil {
			return icon
		}
	}
	icon = fdoLookupApplication(theme, size, win.Command())
	if icon != nil {
		return icon
	}
	return fdoLookupApplication(theme, size, win.IconName())
}

// lookupAnyIconSizeInThemeDir walks file inside of a directory in reverse order to make sure larger sized icons are found first
func lookupAnyIconSizeInThemeDir(directory string, joiner string, iconName string) string {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return ""
	}
	// Lets walk the files in reverse so bigger icons are selected first (Unless it is a 3 digit icon size like 128)
	// Example is /usr/share/icons/icon_theme/<size>/<joiner>/xterm.png
	for i := len(files) - 1; i >= 0; i-- {
		f := files[i]
		if strings.HasPrefix(f.Name(), ".") || !f.IsDir() {
			continue
		}
		matchDir := filepath.Join(directory, f.Name())
		for _, extension := range iconExtensions {
			testIcon := filepath.Join(matchDir, joiner, iconName+extension)
			if _, err := os.Stat(testIcon); err == nil {
				return testIcon
			}
		}
	}

	directory = filepath.Join(directory, joiner)
	files, err = ioutil.ReadDir(directory)
	if err != nil {
		return ""
	}
	// And then walk the directories in <joiner>/<size> order
	// Example is /usr/share/icons/icon_theme/<joiner>/<size>/xterm.png
	for i := len(files) - 1; i >= 0; i-- {
		f := files[i]
		if strings.HasPrefix(f.Name(), ".") || !f.IsDir() {
			continue
		}
		matchDir := filepath.Join(directory, f.Name())
		for _, extension := range iconExtensions {
			testIcon := filepath.Join(matchDir, iconName+extension)
			if _, err := os.Stat(testIcon); err == nil {
				return testIcon
			}
		}
	}

	return ""
}

// lookupIconPathInTheme searches icon locations to find a match using a provided theme directory
func lookupIconPathInTheme(iconSize string, dir string, iconName string) string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return ""
	}
	for _, extension := range iconExtensions {
		// Example is /usr/share/icons/icon_theme/32/apps/xterm.png
		testIcon := filepath.Join(dir, iconSize, "apps", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
		// Example is /usr/share/icons/icon_theme/32x32/apps/xterm.png
		testIcon = filepath.Join(dir, iconSize+"x"+iconSize, "apps", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}

		// Example is /usr/share/icons/icon_theme/apps/32/xterm.png
		testIcon = filepath.Join(dir, "apps", iconSize, iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
		// Example is /usr/share/icons/icon_theme/apps/32x32/xterm.png
		testIcon = filepath.Join(dir, "apps", iconSize+"x"+iconSize, iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}

		// Try this if the requested iconSize didn't exist
		// Example is /usr/share/icons/icon_theme/scalable/apps/xterm.png
		testIcon = filepath.Join(dir, "scalable", "apps", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
		// Example is /usr/share/icons/icon_theme/apps/scalable/xterm.png
		testIcon = filepath.Join(dir, "apps", "scalable", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
	}
	// If the requested icon wasn't found in the specific size or scalable dirs
	// of the theme try all sizes within theme and all icon dirs besides apps
	var subIconDirs = []string{"apps", "actions", "devices", "emblems", "mimetypes", "places", "status"}
	for _, joiner := range subIconDirs {
		testIcon := lookupAnyIconSizeInThemeDir(dir, joiner, iconName)
		if testIcon != "" {
			return testIcon
		}
	}
	return ""
}

//fdoLookupIconPath will take the name of an icon and find a matching image file
func fdoLookupIconPath(theme string, size int, iconName string) string {
	locationLookup := fdoLookupXdgDataDirs()
	iconTheme := theme
	iconSize := fmt.Sprintf("%d", size)
	for _, dataDir := range locationLookup {
		//Example is /usr/share/icons/icon_theme
		dir := filepath.Join(dataDir, "icons", iconTheme)
		iconPath := lookupIconPathInTheme(iconSize, dir, iconName)
		if iconPath != "" {
			return iconPath
		}
	}
	for _, dataDir := range locationLookup {
		//Hicolor is the default fallback theme - Example /usr/share/icons/icon_theme/hicolor
		dir := filepath.Join(dataDir, "icons", "hicolor")
		iconPath := lookupIconPathInTheme(iconSize, dir, iconName)
		if iconPath != "" {
			return iconPath
		}
	}
	for _, dataDir := range locationLookup {
		//Icons may be in the pixmaps directory - test before we do our final fallback
		for _, extension := range iconExtensions {
			iconPath := filepath.Join(dataDir, "pixmaps", iconName+extension)
			if _, err := os.Stat(iconPath); err == nil {
				return iconPath
			}
		}
	}
	for _, dataDir := range locationLookup {
		//Icon was not found in default theme or default fallback theme - Check all themes for a match
		//Example is /usr/share/icons
		files, err := ioutil.ReadDir(filepath.Join(dataDir, "icons"))
		if err != nil {
			continue
		}
		//Enter icon theme
		for _, f := range files {
			if strings.HasPrefix(f.Name(), ".") || !f.IsDir() {
				continue
			}

			//Example is /usr/share/icons/gnome
			iconPath := lookupIconPathInTheme(iconSize, filepath.Join(dataDir, "icons", f.Name()), iconName)
			if iconPath != "" {
				return iconPath
			}
		}
	}
	//No Icon Was Found
	return ""
}

//newFdoIconData creates and returns a struct that contains needed fields from a .desktop file
func newFdoIconData(theme string, size int, desktopPath string) *fdoIconData {
	file, err := os.Open(desktopPath)
	if err != nil {
		fyne.LogError("Could not open file", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fdoApp := fdoIconData{name: "", iconName: "", iconPath: "", exec: ""}
	var currentSection string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "[") {
			currentSection = line
		}
		if currentSection != "[Desktop Entry]" {
			continue
		}
		if strings.HasPrefix(line, "Name=") {
			name := strings.SplitAfter(line, "=")
			fdoApp.name = name[1]
		} else if strings.HasPrefix(line, "Icon=") {
			icon := strings.SplitAfter(line, "=")
			fdoApp.iconName = icon[1]
			if _, err := os.Stat(icon[1]); err == nil {
				fdoApp.iconPath = icon[1]
			} else {
				fdoApp.iconPath = fdoLookupIconPath(theme, size, fdoApp.iconName)
			}
			if fdoApp.iconPath == "" {
				fyne.LogError("Could not find path for icon "+fdoApp.iconName, nil)
			}
		} else if strings.HasPrefix(line, "Exec=") {
			exec := strings.SplitAfter(line, "=")
			fdoApp.exec = exec[1]
		}
	}
	if err := scanner.Err(); err != nil {
		fyne.LogError("Could not read file", err)
		return nil
	}
	return &fdoApp
}

type fdoIconProvider struct {
}

//FindIconFromAppName matches an icon name to a location and returns an IconData interface
func (f *fdoIconProvider) FindIconFromAppName(theme string, size int, appName string) desktop.IconData {
	return fdoLookupApplication(theme, size, appName)
}

//FindIconFromPartialAppName returns a list of icons that match a partial name of an app and returns an IconData slice
func (f *fdoIconProvider) FindIconsMatchingAppName(theme string, size int, appName string) []desktop.IconData {
	return fdoLookupApplicationMatching(theme, size, appName)
}

//FindIconFromWinInfo matches window information to an icon location and returns an IconData interface
func (f *fdoIconProvider) FindIconFromWinInfo(theme string, size int, win desktop.Window) desktop.IconData {
	return fdoLookupApplicationWinInfo(theme, size, win)
}

// NewFDOIconProvider returns a new icon provider following the FreeDesktop.org specifications
func NewFDOIconProvider() desktop.IconProvider {
	return &fdoIconProvider{}
}
