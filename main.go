package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pgaskin/lithiumpatch/dict"
	_ "github.com/pgaskin/lithiumpatch/dict/edgedict"
	_ "github.com/pgaskin/lithiumpatch/dict/webster1913"
	"github.com/pgaskin/lithiumpatch/fonts"
	"github.com/pgaskin/lithiumpatch/patches"
	"github.com/pgaskin/lithiumpatch/patches/patchdef"

	"github.com/spf13/pflag"

	_ "github.com/ncruces/go-sqlite3/embed"
)

var defaultKeySig = "06:7D:43:11:08:F8:ED:AF:24:71:7D:CE:D1:A3:01:D9:55:A3:A2:90"

var (
	Keystore           = pflag.StringP("keystore", "k", "default.jks", "Path to keystore for signing (will be created if does not exist)")
	KeystoreAlias      = pflag.String("keystore-alias", "default", "Keystore alias")
	KeystorePassphrase = pflag.String("keystore-passphrase", "default", "Keystore passphrase")
	Output             = pflag.StringP("output", "o", "", "Output APK path (default: {basename}.patched.resigned.apk)")
	Diff               = pflag.StringP("diff", "d", "", "Write diff to the specified file (default: disabled)")

	AddFonts = pflag.StringSlice("add-fonts", nil, "Add extra TTF fonts from a directory (Regular/Roman, Bold, Italic, and BoldItalic variants should be provided) (can be specified multiple times)")

	Apktool   = pflag.String("apktool", "lib/apktool-2.8.1.jar", "Path to apktool.jar (2.8.1)")
	Apksigner = pflag.String("apksigner", "lib/apksigner-0.9.jar", "Path to apksigner.jar (0.9 or later)")
	Zipalign  = pflag.String("zipalign", "zipalign", "zipalign executable (will search PATH)")
	Keytool   = pflag.String("keytool", "keytool", "keytool executable (will search PATH)")

	Quiet = pflag.BoolP("quiet", "q", false, "Do not show the diff")
	Help  = pflag.Bool("help", false, "Show this help text")
)

func main() {
	pflag.CommandLine.SortFlags = false
	pflag.Parse()

	if *Help || pflag.NArg() != 1 {
		fmt.Printf("usage: %s [options] APK_PATH\n\noptions:\n%s", os.Args[0], pflag.CommandLine.FlagUsages())
		os.Exit(1)
	}

	fmt.Printf("> Loading extra fonts\n")
	for _, x := range *AddFonts {
		n, err := fonts.LoadFrom(os.DirFS(x))
		if n == 0 {
			err = fmt.Errorf("no fonts found")
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: add fonts from %q: %v\n", x, err)
			os.Exit(1)
		}
	}
	for _, x := range fonts.All() {
		fmt.Printf("... %s\n", x)
	}
	fmt.Println()

	fmt.Printf("> Parsing dictionaries\n")
	if err := dict.Parse(true); err != nil {
		fmt.Fprintf(os.Stderr, "error: parse dictionaries: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	apk := pflag.Arg(0)

	if _, err := os.Stat(apk); err != nil {
		return err
	}

	if *Output == "" {
		*Output = strings.TrimSuffix(apk, filepath.Ext(apk)) + ".patched.resigned.apk"
	}

	fmt.Printf("> Creating temp dirs\n")
	tmp, err := os.MkdirTemp("", "lithiumpatch")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer cleanup(tmp, true)

	apkTmpDir := filepath.Join(tmp, "apk")
	if err := os.Mkdir(apkTmpDir, 0777); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	disTmpDir := filepath.Join(tmp, "dis")
	if err := os.Mkdir(disTmpDir, 0777); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	fmt.Println()

	fmt.Printf("> Looking for keystore %q\n", *Keystore)
	if _, err := os.Stat(*Keystore); os.IsNotExist(err) {
		fmt.Printf("> Generating keystore %q\n", *Keystore)
		cmd := exec.CommandContext(ctx,
			*Keytool, "-genkeypair", "-v",
			"-keystore", *Keystore,
			"-keyalg", "RSA",
			"-storepass", *KeystorePassphrase,
			"-alias", *KeystoreAlias,
			"-validity", "3652",
			"-dname", "CN=lithiumpatch",
		)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("keytool: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("access keystore: %w", err)
	}
	fmt.Println()

	fmt.Printf("> Reading keystore %q\n", *Keystore)
	var b bytes.Buffer
	cmd := exec.Command(
		"keytool", "-list", "-v",
		"-keystore", *Keystore,
		"-keyalg", "RSA",
		"-storepass", *KeystorePassphrase,
		"-alias", *KeystoreAlias,
	)
	cmd.Stdout = &b
	cmd.Stderr = &b
	if err := cmd.Run(); err != nil {
		fmt.Println(b.String())
		return fmt.Errorf("keytool: %w", err)
	}
	m := regexp.MustCompile(`SHA1: *([A-Fa-f0-9:]+)`).FindStringSubmatch(b.String())
	if len(m) != 2 {
		fmt.Println(b.String())
		return fmt.Errorf("could not find fingerprint in keytool output")
	}
	if m[1] == defaultKeySig {
		patches.NoSync()
		fmt.Fprintf(os.Stderr, "Found default signing key. This is insecure and will not support sync. You can specify a custom keystore using the --keystore option.\n")
	} else {
		fmt.Fprintf(os.Stderr, "Found key with signature %s. This will need to be added as Google APIs app with access to the Drive API for sync to work.\n", m[1])
	}
	fmt.Println()

	fmt.Printf("> Decompiling APK %q to %q\n", apk, disTmpDir)
	if err := jar(ctx, *Apktool, "d", "-f", apk, "-o", disTmpDir); err != nil {
		return fmt.Errorf("apktool: %w", err)
	}
	fmt.Println()

	fmt.Printf("> Patching\n")
	diff := new(bytes.Buffer)
	ps := patchdef.Patches()
	for i, patch := range ps {
		fmt.Printf("[%d/%d] %s\n", i+1, len(ps), patch.Name())
		if err := patch.Apply(disTmpDir, diff); err != nil {
			return fmt.Errorf("apply patch %q: %w", patch.Name(), err)
		}
	}
	if !*Quiet {
		fmt.Println(diff.String())
	}
	if *Diff != "" {
		if err := os.WriteFile(*Diff, diff.Bytes(), 0666); err != nil {
			return fmt.Errorf("write diff: %w", err)
		}
	}
	fmt.Println()

	apkPatched := filepath.Join(apkTmpDir, "patched.apk")
	fmt.Printf("> Compiling APK to %q\n", apkPatched)
	if err := jar(ctx, *Apktool, "b", "-f", disTmpDir, "-o", apkPatched); err != nil {
		return fmt.Errorf("apktool: %w", err)
	}
	fmt.Println()

	fmt.Printf("> Zipaligning APK\n")
	apkPatchedBeforeAlign := filepath.Join(apkTmpDir, "unaligned.apk")
	if err := os.Rename(apkPatched, apkPatchedBeforeAlign); err != nil {
		return fmt.Errorf("rename apk: %w", err)
	}
	cmd = exec.CommandContext(ctx, *Zipalign, "4", filepath.Join(apkTmpDir, "unaligned.apk"), apkPatched)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("zipalign: %w", err)
	}
	fmt.Println()

	fmt.Printf("> Signing APK %q to %q\n", apkPatched, *Output)
	if err := jar(ctx,
		*Apksigner, "sign",
		"--min-sdk-version", "26",
		"--ks", *Keystore,
		"--ks-pass", "pass:"+*KeystorePassphrase,
		"--pass-encoding", "utf-8",
		"--ks-key-alias", *KeystoreAlias,
		"--out", *Output,
		apkPatched,
	); err != nil {
		return fmt.Errorf("apksigner: %w", err)
	}
	fmt.Println()

	fmt.Println("done")
	return nil
}

func cleanup(path string, all bool) {
	fmt.Printf("> Cleaning up %s\n", path)
	var err error
	if all {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}
	if err != nil {
		fmt.Printf("Warning: %v", err)
	}
}

func jar(ctx context.Context, jarfile string, args ...string) error {
	if _, err := exec.LookPath("java"); err != nil {
		return fmt.Errorf("could not find java: %v", err)
	}

	if _, err := os.Stat(jarfile); err != nil {
		return fmt.Errorf("could not access jarfile %q: %v", jarfile, err)
	}

	cmd := exec.CommandContext(ctx, "java", append([]string{"-jar", jarfile}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
