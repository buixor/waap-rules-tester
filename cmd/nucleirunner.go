package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type TestSuite struct {
	Name            string   `yaml:"name"`
	NucleiTemplates []string `yaml:"nuclei_templates"`
	FailureSeverity string   `yaml:"failure_severity"` //high | medium | low
}

type TestSuiteResult struct {
	Suite            TestSuite
	FailedTemplates  []string
	SuccessTemplates []string
	ErroredTemplates []string
}

func LoadTestSuite(path string) (TestSuite, error) {
	var test TestSuite

	file, err := os.ReadFile(path)
	if err != nil {
		return test, err
	}
	if err := yaml.UnmarshalStrict(file, &test); err != nil {
		return test, err
	}
	if len(test.NucleiTemplates) == 0 {
		return test, fmt.Errorf("No nuclei templates specified")
	}
	if test.Name == "" {
		return test, fmt.Errorf("No name specified")
	}
	if test.FailureSeverity == "" {
		test.FailureSeverity = "medium"
	}
	return test, nil
}

var NucleiTemplateFail = errors.New("Nuclei template failed")

func (ts *TestSuite) RunNucleiTemplate(template_path string, tstamp int64) error {
	//template_path is the full path to the template, we just want the name ie. "sqli-random-test"
	tmp := strings.Split(template_path, "/")
	template := strings.Split(tmp[len(tmp)-1], ".")[0]
	tmpFile := fmt.Sprintf("%s/%s_%s-%d.json", cfg.TempDir, ts.Name, template, tstamp)
	args := []string{
		"-u", cfg.Target,
		"-t", template_path,
		"-o", tmpFile,
	}
	args = append(args, cfg.NucleiConfig.NucleiOptions...)
	cmd := exec.Command(cfg.NucleiConfig.NucleiPath, args...)

	var out bytes.Buffer
	var out_err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out_err
	err := cmd.Run()
	stdout_fname := fmt.Sprintf("%s/%s-%s-%d_stdout.txt", cfg.TempDir, template, ts.Name, tstamp)
	if err := os.WriteFile(stdout_fname, out.Bytes(), 0644); err != nil {
		log.Warningf("Error writing stdout: %s", err)
	}
	stderr_fname := fmt.Sprintf("%s/%s-%s-%d_stderr.txt", cfg.TempDir, template, ts.Name, tstamp)
	if err := os.WriteFile(stderr_fname, out_err.Bytes(), 0644); err != nil {
		log.Warningf("Error writing stderr: %s", err)
	}
	if err != nil {
		log.Warningf("Error running nuclei: %s", err)
		log.Warningf("Stdout saved to %s", stdout_fname)
		log.Warningf("Stderr saved to %s", stderr_fname)
		log.Warningf("Nuclei generated output saved to %s", tmpFile)
		return err
	} else if len(out.String()) == 0 {
		//No stdout means no finding, it means our test failed
		return NucleiTemplateFail
	}
	return nil
}

func (ts *TestSuite) RunTestSuite(suite TestSuite) (TestSuiteResult, error) {
	results := TestSuiteResult{Suite: suite}
	test_ts := time.Now().Unix()
	for idx, template := range suite.NucleiTemplates {
		log.Infof("Running test %d/%d of %s", idx+1, len(suite.NucleiTemplates), template)
		if err := ts.RunNucleiTemplate(template, test_ts); err != nil {
			if err == NucleiTemplateFail {
				log.Warningf("Nuclei template %s failed, continuing", template)
				results.FailedTemplates = append(results.FailedTemplates, template)
			} else {
				log.Errorf("Aborting test suite due to error: %s", err)
				results.ErroredTemplates = append(results.ErroredTemplates, template)
			}
		} else {
			results.SuccessTemplates = append(results.SuccessTemplates, template)
		}
	}
	return results, nil
}
