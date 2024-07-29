    return memBytesTotal / (double) numTasks;
  }

  public double getAvgCpuUsed() {
    return cpuTotal / (double) numTasks;
  }

  public String getDeployId() {
    return deployId;
  }

  public String getRequestId() {
    return requestId;
  }
}
