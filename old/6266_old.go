	if err != nil {
		return err
	}
	return nil
}

// ApplyTerraformModule does an equivalent of `terraform apply -auto-approve`
func (tm *TerraformModule) ApplyTerraformModule() error {
	modulePathDir := filepath.Dir(tm.ModulePath)
	modulePathDirName := filepath.Base(modulePathDir)

	// TODO(prs): Replace this with stable terraform-exec when it is released
	// cd {modulePathDirName} && terraform init && terraform apply -auto-approve
	//
	// e.g.
	// cd deploy/production/ws-us89/terraform && terraform init && terraform apply -auto-approve
	commandToRun := fmt.Sprintf(`cd %s && terraform init && terraform apply -auto-approve`, modulePathDirName)

	cmd := exec.Command("/bin/sh", "-c", commandToRun)
	err := cmd.Run()
	stdout, _ := cmd.Output()
	fmt.Print(string(stdout))
	if err != nil {
		return err
	}
	return nil
}

// DestroyTerraformModule does an equivalent of `terraform destroy -auto-approve`
func (tm *TerraformModule) DestroyTerraformModule() error {
	modulePathDir := filepath.Dir(tm.ModulePath)
	modulePathDirName := filepath.Base(modulePathDir)

	// TODO(prs): Replace this with stable terraform-exec when it is released
	// cd {modulePathDirName} && terraform destroy -auto-approve
	//
	// e.g.
	// cd deploy/production/ws-us89/terraform && terraform destroy -auto-approve
	commandToRun := fmt.Sprintf(`cd %s && terraform destroy -auto-approve`, modulePathDirName)

	cmd := exec.Command("/bin/sh", "-c", commandToRun)
	err := cmd.Run()
	stdout, _ := cmd.Output()
	fmt.Print(string(stdout))
	if err != nil {
		return err
	}
	return nil
}
