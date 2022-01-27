package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestKnowledgeBase(t *testing.T) {
	kbFolder := os.Getenv("KBFolder")

	if kbFolder == "" {
		kbFolder = "knowledge-base"
	}

	lintIssues := []string{}

	err := filepath.Walk(kbFolder,
		func(filePath string, info os.FileInfo, err error) error {
			if !strings.HasSuffix(info.Name(), "yml") && !strings.HasSuffix(info.Name(), "yaml") {
				return nil
			}

			if err != nil {
				lintIssues = append(lintIssues, fmt.Sprintf("Error reading %s: %v", filePath, err))
				return nil
			}
			if info.Name() != "action-security.yml" {
				lintIssues = append(lintIssues, fmt.Sprintf("File must be named action-security.yml, not %s at %s", info.Name(), filePath))
				return nil
			}
			input, err := ioutil.ReadFile(filePath)

			if err != nil {
				lintIssues = append(lintIssues, fmt.Sprintf("Unable to read action-security.yml at %s", filePath))
				return nil
			}

			actionMetadata := ActionMetadata{}

			err = yaml.Unmarshal([]byte(input), &actionMetadata)
			if err != nil {
				lintIssues = append(lintIssues, fmt.Sprintf("Unable to unmarshall action-security.yml at %s", filePath))
				return nil
			}

			if actionMetadata.Name == "" {
				lintIssues = append(lintIssues, fmt.Sprintf("Name must not be empty in action-security.yml at %s", filePath))
				return nil
			}

			for _, endpoint := range actionMetadata.AllowedEndpoints {
				if endpoint.FQDN == "" {
					lintIssues = append(lintIssues, fmt.Sprintf("FQDN must not be empty in action-security.yml at %s", filePath))
					return nil
				}

				if strings.ToLower(endpoint.FQDN) != endpoint.FQDN {
					lintIssues = append(lintIssues, fmt.Sprintf("FQDN must be all lower case in action-security.yml at %s", filePath))
					return nil
				}

				if endpoint.Port == 0 {
					lintIssues = append(lintIssues, fmt.Sprintf("Port must not be empty in action-security.yml at %s", filePath))
					return nil
				}

				if endpoint.Reason == "" {
					lintIssues = append(lintIssues, fmt.Sprintf("Reason must not be empty for fqdn %s in action-security.yml at %s", endpoint.FQDN, filePath))
					return nil
				}

				if !strings.HasPrefix(endpoint.Reason, "to ") {
					lintIssues = append(lintIssues, fmt.Sprintf("Reason must start with 'to '. It is currently %s in action-security.yml at %s", endpoint.Reason, filePath))
					return nil
				}

				if strings.HasSuffix(endpoint.Reason, ".") {
					lintIssues = append(lintIssues, fmt.Sprintf("Reason must not end with '.'. It is currently %s in action-security.yml at %s", endpoint.Reason, filePath))
					return nil
				}
			}

			validScopes := []string{"actions", "checks", "contents", "deployments", "id-token", "issues", "packages",
				"pull-requests", "repository-projects", "security-events", "statuses"}
			mapScopes := make(map[string]bool)

			for _, scope := range validScopes {
				mapScopes[scope] = true
			}

			for key, scope := range actionMetadata.GitHubToken.Permissions.Scopes {

				_, found := mapScopes[key]
				if !found {
					lintIssues = append(lintIssues, fmt.Sprintf("Scope must be one of %s. It is currently %s in action-security.yml at %s", strings.Join(validScopes, ","), key, filePath))
					return nil
				}

				if scope.Permission != "read" && scope.Permission != "write" {
					lintIssues = append(lintIssues, fmt.Sprintf("Permissions must be either read or write. It is currently %s in action-security.yml at %s", scope.Permission, filePath))
					return nil
				}

				if !strings.HasPrefix(scope.Reason, "to ") {
					lintIssues = append(lintIssues, fmt.Sprintf("Reason must start with 'to '. It is currently %s in action-security.yml at %s", scope.Reason, filePath))
					return nil
				}

				if strings.HasSuffix(scope.Reason, ".") {
					lintIssues = append(lintIssues, fmt.Sprintf("Reason must not end with '.'. It is currently %s in action-security.yml at %s", scope.Reason, filePath))
					return nil
				}

				//Since the reason is added as a comment in the workflow file, limit the length to 50 to not clutter the workflow file
				if len(scope.Reason) > 50 {
					lintIssues = append(lintIssues, fmt.Sprintf("Reason must not exceed 50 char limit. It is currently %d in action-security.yml at %s", len(scope.Reason), filePath))
				}
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	if len(lintIssues) > 0 {
		for _, issue := range lintIssues {
			log.Println(issue)
		}
		t.Fail()
	}
}
