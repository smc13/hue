package hue

import (
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
)

// filterStacktraceFrames filters the stack trace frames to only include those that are relevant
// ignoring frames from go runtime and other internal packages.
func filterStacktraceFrames(frames *runtime.Frames) []runtime.Frame {
	filtered := make([]runtime.Frame, 0)
	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		filtered = append(filtered, frame)
	}

	return filtered
}

// ensure package path and parentheses are removed from function names
// turns "/path/to/package.funcName" into "funcName"
// turns "/path/to/package.(*Type).Method" into "Type.Method"
func formatFunctionName(function string) string {
	withoutPath := filepath.Base(function)

	// remove everything before the first period
	if dotIndex := strings.Index(withoutPath, "."); dotIndex != -1 {
		withoutPath = withoutPath[dotIndex+1:]
	}

	// remove ( and ) and *
	withoutPath = strings.ReplaceAll(withoutPath, "(", "")
	withoutPath = strings.ReplaceAll(withoutPath, ")", "")
	withoutPath = strings.ReplaceAll(withoutPath, "*", "")

	return withoutPath
}

// cleanFramePath cleans up the file path in the stack trace frame.
// It removes the module path if the file is part of a Go module or standard library.
func cleanFramePath(file string) string {
	absFile := filepath.ToSlash(file)

	// check if the file is part of the main module (e.g., in a Go project)
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			// vcs.checkoutDir should be the most reliable way to find the main module path
			if setting.Key == "vcs.checkoutDir" && setting.Value != "" {
				prefix := filepath.ToSlash(setting.Value)
				if strings.HasPrefix(absFile, prefix) {
					return strings.TrimPrefix(absFile, prefix+"/")
				}
			}

			// fallback for older Go versions
			if setting.Key == "mod" && setting.Value != "" {
				prefix := filepath.ToSlash(setting.Value)
				if strings.HasPrefix(absFile, prefix) {
					return strings.TrimPrefix(absFile, prefix+"/")
				}
			}
		}

		// could be a dependency
		for _, dep := range info.Deps {
			if dep.Replace != nil {
				dep = dep.Replace
			}
			modPath := filepath.ToSlash(dep.Path)
			version := dep.Version
			pattern := modPath + "@" + version
			if idx := strings.Index(absFile, pattern); idx != -1 {
				// Strip up to module@version/
				rel := absFile[idx+len(pattern)+1:]
				return modPath + "/" + rel
			}
		}
	}

	// we can also check if the file is part of the Go standard library
	goRoot := filepath.ToSlash(runtime.GOROOT())
	if strings.HasPrefix(absFile, goRoot+"/src/") {
		return strings.TrimPrefix(absFile, goRoot+"/src/")
	}

	return absFile
}
