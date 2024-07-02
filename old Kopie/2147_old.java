package com.hubspot.singularity;
  DEPLOY

public enum SingularityAuthorizationScope {
  READ,
  WRITE,
  ADMIN,
  DEPLOY,
  EXEC // run, bounce, kill, pause, scale
}
