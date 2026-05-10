package tools

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

func GetContexts(input map[string]string) (string, error) {
	cmd := exec.Command(
		"kubectl",
		"config",
		"get-contexts",
		"-o",
		"name",
	)

	out, err := cmd.CombinedOutput()
	return string(out), err
}

func SetContexts(input map[string]string) (string, error) {
	c, ok := input["context"]
	if !ok || c == "" {
		return "", fmt.Errorf("context is required")
	}

	// Get available contexts
	getCmd := exec.Command(
		"kubectl",
		"config",
		"get-contexts",
		"-o=name",
	)

	contextsOut, err := getCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get contexts: %w", err)
	}

	// Validate context exists
	contexts := strings.Split(strings.TrimSpace(string(contextsOut)), "\n")

	found := false
	for _, ctx := range contexts {
		if ctx == c {
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("context %q does not exist", c)
	}
	cmd := exec.Command(
		"kubectl",
		"config",
		"set-context",
		c,
	)

	out, err := cmd.CombinedOutput()
	return string(out), err
}

func GetPods(input map[string]string) (string, error) {
	ns := input["namespace"]
	cmd := exec.Command(
		"kubectl",
		"get",
		"pods",
		"-n",
		ns,
		"-o",
		"json",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	var podList K8sPodList
	if err := json.Unmarshal(out, &podList); err != nil {
		return "", err
	}

	var summaries []PodSummary

	for _, p := range podList.Items {
		summary := PodSummary{
			Name:      p.Metadata.Name,
			Namespace: p.Metadata.Namespace,
			Phase:     p.Status.Phase,
		}

		if len(p.Status.ContainerStatuses) > 0 {
			cs := p.Status.ContainerStatuses[0]
			summary.Ready = cs.Ready
			summary.Restarts = cs.RestartCount
			summary.Reason = cs.State.Waiting.Reason
		}

		summaries = append(summaries, summary)
	}

	result, err := json.MarshalIndent(summaries, "", "  ")
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func GetNamespaces(input map[string]string) (string, error) {
	cmd := exec.Command(
		"kubectl",
		"get",
		"ns",
		"-o",
		"json",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	var nsList K8sNamespacesList
	if err := json.Unmarshal(out, &nsList); err != nil {
		return "", err
	}
	var namespaces []NamespaceSummary

	for _, ns := range nsList.Items {
		summary := NamespaceSummary{
			Namespace: ns.Metadata.Name,
		}

		namespaces = append(namespaces, summary)
	}

	result, err := json.MarshalIndent(namespaces, "", "  ")
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func DescribePod(input map[string]string) (string, error) {
	namespace := input["namespace"]
	pod := input["pod"]
	cmd := exec.Command(
		"kubectl",
		"describe",
		"pod",
		pod,
		"-n",
		namespace,
	)
	fmt.Println(cmd)
	fmt.Println(pod)
	fmt.Println(namespace)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func GetPodLogs(input map[string]string) (string, error) {
	namespace := input["namespace"]
	pod := input["pod"]
	cmd := exec.Command(
		"kubectl",
		"logs",
		pod,
		"-n",
		namespace,
		"--tail=300",
		"--all-containers=true",
	)

	out, err := cmd.CombinedOutput()
	return string(out), err
}
