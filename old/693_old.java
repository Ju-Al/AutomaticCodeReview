package com.hubspot.singularity.executor.task;

      templateManager.writeDockerScript(getPath("runner.sh"), new DockerContext(environmentContext, runnerContext, configuration.getDockerPrefix(), configuration.getDockerStopTimeout()));
import java.nio.file.Path;
import java.util.List;
import java.util.concurrent.Callable;

import org.apache.mesos.Protos.TaskInfo;
import org.apache.mesos.Protos.TaskState;

import com.google.common.base.Optional;
import com.google.common.collect.Lists;
import com.hubspot.deploy.ExecutorData;
import com.hubspot.singularity.executor.TemplateManager;
import com.hubspot.singularity.executor.config.SingularityExecutorConfiguration;
import com.hubspot.singularity.executor.models.DockerContext;
import com.hubspot.singularity.executor.models.EnvironmentContext;
import com.hubspot.singularity.executor.models.RunnerContext;
import com.hubspot.singularity.executor.task.SingularityExecutorArtifactFetcher.SingularityExecutorTaskArtifactFetcher;
import com.hubspot.singularity.executor.utils.ExecutorUtils;
import com.spotify.docker.client.DockerClient;

public class SingularityExecutorTaskProcessBuilder implements Callable<ProcessBuilder> {

  private final SingularityExecutorTask task;

  private final TemplateManager templateManager;
  private final SingularityExecutorConfiguration configuration;

  private final String executorPid;

  private final ExecutorUtils executorUtils;

  private final ExecutorData executorData;

  private final SingularityExecutorArtifactFetcher artifactFetcher;

  private Optional<SingularityExecutorTaskArtifactFetcher> taskArtifactFetcher;

  private DockerClient dockerClient;

  public SingularityExecutorTaskProcessBuilder(SingularityExecutorTask task,
                                               ExecutorUtils executorUtils,
                                               SingularityExecutorArtifactFetcher artifactFetcher,
                                               TemplateManager templateManager,
                                               SingularityExecutorConfiguration configuration,
                                               ExecutorData executorData, String executorPid,
                                               DockerClient dockerClient) {
    this.executorData = executorData;
    this.task = task;
    this.executorUtils = executorUtils;
    this.artifactFetcher = artifactFetcher;
    this.templateManager = templateManager;
    this.configuration = configuration;
    this.executorPid = executorPid;
    this.taskArtifactFetcher = Optional.absent();
    this.dockerClient = dockerClient;
  }

  @Override
  public ProcessBuilder call() throws Exception {
    if (task.getTaskInfo().hasContainer() && task.getTaskInfo().getContainer().hasDocker()) {
      executorUtils.sendStatusUpdate(task.getDriver(), task.getTaskInfo(), TaskState.TASK_STARTING, String.format("Pulling image... (executor pid: %s)", executorPid), task.getLog());
      dockerClient.pull(task.getTaskInfo().getContainer().getDocker().getImage());
    }

    executorUtils.sendStatusUpdate(task.getDriver(), task.getTaskInfo(), TaskState.TASK_STARTING, String.format("Staging files... (executor pid: %s)", executorPid), task.getLog());

    taskArtifactFetcher = Optional.of(artifactFetcher.buildTaskFetcher(executorData, task));
    taskArtifactFetcher.get().fetchFiles();

    ProcessBuilder processBuilder = buildProcessBuilder(task.getTaskInfo(), executorData);

    task.getTaskLogManager().setup();

    return processBuilder;
  }

  public void cancel() {
    if (taskArtifactFetcher.isPresent()) {
      taskArtifactFetcher.get().cancel();
    }
  }

  private Path getPath(String filename) {
    return task.getTaskDefinition().getTaskDirectoryPath().resolve(filename);
  }

  private String getCommand(ExecutorData executorData) {
    final StringBuilder bldr = new StringBuilder(executorData.getCmd());
    for (String extraCmdLineArg : executorData.getExtraCmdLineArgs()) {
      bldr.append(" ");
      bldr.append(extraCmdLineArg);
    }
    return bldr.toString();
  }

  private String getExecutorUser() {
    return System.getProperty("user.name");  // TODO: better way to do this?
  }

  private ProcessBuilder buildProcessBuilder(TaskInfo taskInfo, ExecutorData executorData) {
    final String cmd = getCommand(executorData);

    RunnerContext runnerContext = new RunnerContext(
      cmd,
      configuration.getTaskAppDirectory(),
      configuration.getLogrotateToDirectory(),
      executorData.getUser().or(configuration.getDefaultRunAsUser()),
      configuration.getServiceLog(),
      task.getTaskId(),
      executorData.getMaxTaskThreads().or(configuration.getMaxTaskThreads()),
      !getExecutorUser().equals(executorData.getUser().or(configuration.getDefaultRunAsUser())));
    EnvironmentContext environmentContext = new EnvironmentContext(taskInfo);
    if (taskInfo.hasContainer() && taskInfo.getContainer().hasDocker()) {
      task.getLog().info("Writing a runner script to execute {} in docker container", cmd);
      boolean privileged = (taskInfo.getContainer().getDocker().hasPrivileged() && taskInfo.getContainer().getDocker().getPrivileged());
      templateManager.writeDockerScript(getPath("runner.sh"), new DockerContext(environmentContext, runnerContext, configuration.getDockerPrefix(), configuration.getDockerStopTimeout(), privileged));
    } else {
      templateManager.writeEnvironmentScript(getPath("deploy.env"), environmentContext);

      task.getLog().info("Writing a runner script to execute {} with {}", cmd, runnerContext);

      templateManager.writeRunnerScript(getPath("runner.sh"), runnerContext);
    }

    List<String> command = Lists.newArrayList();
    command.add("bash");
    command.add("runner.sh");

    ProcessBuilder processBuilder = new ProcessBuilder(command);

    processBuilder.directory(task.getTaskDefinition().getTaskDirectoryPath().toFile());

    processBuilder.redirectError(task.getTaskDefinition().getExecutorBashOutPath().toFile());
    processBuilder.redirectOutput(task.getTaskDefinition().getExecutorBashOutPath().toFile());

    return processBuilder;
  }

  @Override
  public String toString() {
    return "SingularityExecutorTaskProcessBuilder [task=" + task.getTaskId() + "]";
  }

}
