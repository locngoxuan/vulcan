package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/locngoxuan/vulcan/core"
)

//curl -u 'admin:123456a@!' -X PUT 'http://localhost:8081/artifactory/wes-libs-snapshot-local/com/fortna/athena/database/1.5.0-SNAPSHOT/database-1.5.0-SNAPSHOT.tar'
// -T '/mnt/hgfs/workspace/athena/database/build/database-1.5.0-SNAPSHOT.tar' -H 'X-CheckSum-MD5:c3e49b7305b07a3d64f8a3d667a598dc'
func main() {
	repository := flag.String("repository", "", "specify repository")
	group := flag.String("group", "", "specify group of artifact")
	artifactId := flag.String("artifact", "", "specify artifact id")
	version := flag.String("version", "", "specify version of artifact")
	inSecure := flag.Bool("insecure", false, "allow connections to SSL sites without certs")
	username := flag.String("username", "", "username of jfrog's account")
	password := flag.String("password", "", "password of jfrog's account")
	source := flag.String("source", "", "specify location of source")
	flag.Parse()

	if *group = strings.TrimSpace(*group); *group == "" {
		fmt.Println("group is missing")
		os.Exit(1)
	}

	if *artifactId = strings.TrimSpace(*artifactId); *artifactId == "" {
		fmt.Println("artifact id is missing")
		os.Exit(1)
	}

	if *repository = strings.TrimSpace(*repository); *repository == "" {
		fmt.Println("repository is missing")
		os.Exit(1)
	}

	if *username = strings.TrimSpace(*username); *username == "" {
		fmt.Println("username is missing")
		os.Exit(1)
	}

	if *password = strings.TrimSpace(*password); *password == "" {
		fmt.Println("password is missing")
		os.Exit(1)
	}

	if *source = strings.TrimSpace(*source); *source == "" {
		fmt.Println("source is missing")
		os.Exit(1)
	}

	if !isValid(groupPattern, *group) {
		fmt.Println(`group value is malformed`)
		os.Exit(1)
	}

	if !isValid(artifactPattern, *artifactId) {
		fmt.Println(`artifact value is malformed`)
		os.Exit(1)
	}

	fi, err := os.Stat(*source)
	if err != nil {
		fmt.Printf(`source is malformed: %s\n`, err.Error())
		os.Exit(1)
	}
	if fi.IsDir() {
		fmt.Println(`source is not file`)
		os.Exit(1)
	}

	artPkg := ArtifactoryPackage{
		Repository: *repository,
		InSecure:   *inSecure,
		Group:      *group,
		ArtifactId: *artifactId,
		Version:    *version,
		Username:   *username,
		Password:   *password,
		Source:     *source,
	}
	err = artPkg.uploadFile(context.Background())
	if err != nil {
		fmt.Printf("failed to update artifact to repository: %s\n", err.Error())
		os.Exit(1)
	}
}

const groupPattern = `^[a-zA-Z][a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)*[0-9a-zA-Z]$`
const artifactPattern = `^[a-zA-Z]([a-zA-Z0-9_-])*[0-9a-zA-Z]$`

func isValid(ptrn, s string) bool {
	regx := regexp.MustCompile(ptrn)
	if !regx.MatchString(s) {
		return false
	}
	return true
}

type ArtifactoryPackage struct {
	Repository string
	InSecure   bool

	Group      string
	ArtifactId string
	Version    string
	Source     string

	Username string
	Password string
}

func (ap ArtifactoryPackage) uploadFile(ctx context.Context) error {
	fmt.Printf("publish package %s:%s-%s to %s\n", ap.Group, ap.ArtifactId, ap.Version, ap.Repository)
	u, err := url.Parse(ap.Repository)
	if err != nil {
		return err
	}

	//build endpoint
	group := strings.ReplaceAll(ap.Group, ".", "/")
	ext := filepath.Ext(ap.Source)
	fileName := fmt.Sprintf(`%s-%s%s`, ap.ArtifactId, ap.Version, ext)
	var builder strings.Builder
	builder.WriteString(ap.Repository)
	if !strings.HasSuffix(ap.Repository, "/") {
		builder.WriteString("/")
	}
	builder.WriteString(group)
	builder.WriteString("/")
	builder.WriteString(ap.ArtifactId)
	builder.WriteString("/")
	builder.WriteString(ap.Version)
	builder.WriteString("/")
	builder.WriteString(fileName)
	endpoint := path.Clean(builder.String())

	data, err := os.Open(ap.Source)
	if err != nil {
		return err
	}
	defer func() {
		_ = data.Close()
	}()

	md5, err := core.SumContentMD5(ap.Source)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-CheckSum-MD5", md5)
	req.SetBasicAuth(ap.Username, ap.Password)

	client := http.Client{
		Timeout: 60 * time.Second,
	}

	if u.Scheme == "https" {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: ap.InSecure,
			},
		}
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf(res.Status)
	}
	return nil
}
