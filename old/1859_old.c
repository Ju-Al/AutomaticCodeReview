	batch_wrapper_cmd(wrapper, cmd);
	free(cmd);

	cmd = batch_wrapper_write(wrapper, t);
	if(cmd){
		batch_task_set_command(t, cmd);
		struct dag_file *df = makeflow_hook_add_input_file(n->d, t, cmd, cmd, DAG_FILE_TYPE_TEMP);
		debug(D_MAKEFLOW_HOOK, "Wrapper written to %s", df->filename);
		makeflow_log_file_state_change(n->d, df, DAG_FILE_STATE_EXISTS);
	} else {
		debug(D_MAKEFLOW_HOOK, "Failed to create wrapper: errno %d, %s", errno, strerror(errno));
		return MAKEFLOW_HOOK_FAILURE;
	}
	free(cmd);

	makeflow_hook_add_input_file(n->d, t, singularity_image, NULL, DAG_FILE_TYPE_GLOBAL);

	return MAKEFLOW_HOOK_SUCCESS;
}
	
struct makeflow_hook makeflow_hook_singularity = {
	.module_name = "Singularity",
	.create = create,

	.node_submit = node_submit,
};


