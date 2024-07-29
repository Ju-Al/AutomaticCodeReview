    else if (strcmp(arg[iarg],"final_integrate") == 0) final_integrate_flag = 1;
    else if (strcmp(arg[iarg],"final_integrate") == 0) final_integrate_flag = 1;
    else if (strcmp(arg[iarg],"final_integrate") == 0) final_integrate_flag = 1;
    else error->all(FLERR,"Illegal fix DUMMY command");
    iarg++;
  }
}

/* ---------------------------------------------------------------------- */

int FixDummy::setmask()
{
  int mask = 0;
  if (initial_integrate_flag) mask |= INITIAL_INTEGRATE;
  if (final_integrate_flag) mask |= FINAL_INTEGRATE;
  if (pre_exchange_flag) mask |= PRE_EXCHANGE;
  if (pre_neighbor_flag) mask |= PRE_NEIGHBOR;
  if (pre_force_flag) mask |= PRE_FORCE;
  if (post_force_flag) mask |= POST_FORCE;
  if (end_of_step_flag) mask |= END_OF_STEP;
  return mask;
}
