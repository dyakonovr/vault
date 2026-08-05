package logger

import "log"

func Print(args ...any) {
	log.Print(args...)
}

func Printf(format string, args ...any) {
	log.Printf(format, args...)
}

func Error(args ...any) {
	log.Print(args...)
}

func Errorf(format string, args ...any) {
	log.Printf(format, args...)
}

func Fatal(args ...any) {
	log.Fatal(args...)
}

func Fatalf(format string, args ...any) {
	log.Fatalf(format, args...)
}

// func Warn(args ...any) {
// 	log.Warn(args...)
// }

// func Warnf(format string, args ...any) {
// 	log.Warnf(format, args...)
// }

func Info(args ...any) {
	log.Print(args...)
}
