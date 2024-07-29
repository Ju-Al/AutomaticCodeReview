            MPI_Finalize();
            return 1;
        }//corrupted string or something
    }

    //end major 

    char* makeflow_args = "";
    char* workqueue_args = "";
    char* port = "9000";
    char* cpout = NULL;
    char* debug_base = NULL;
    while ((c = getopt_long(argc, argv, "m:q:p:c:d:h", long_options, 0)) != -1) {
        switch (c) {
            case 'm': //makeflow-options
                makeflow_args = xxstrdup(optarg);
                break;
            case 'q': //workqueue-options
                workqueue_args = xxstrdup(optarg);
                break;
            case 'p':
                port = xxstrdup(optarg);
                break;
            case 'h':
                print_help();
                break;
            case 'c':
                cpout = xxstrdup(optarg);
                break;
            case 'd':
                debug_base = xxstrdup(optarg);
                fprintf(stderr, "debug_base: %s\n", debug_base);
                break;
            default: //ignore anything not wanted
                break;
        }
    }

    if (mpi_rank == 0) { //we're makeflow
        char* master_ipaddr = get_ipaddr();
        unsigned mia_len = strlen(master_ipaddr);
        long long value;
        char* key;
        fprintf(stderr, "master ipaddress: %s\n", master_ipaddr);
        hash_table_firstkey(comps);
        while (hash_table_nextkey(comps, &key, (void**) &value)) {
            fprintf(stderr, "sending my ip to %s rank %lli\n", key, value);
            MPI_Send(&mia_len, 1, MPI_UNSIGNED, value, 0, MPI_COMM_WORLD);
            MPI_Send(master_ipaddr, mia_len, MPI_CHAR, value, 0, MPI_COMM_WORLD);
        }
        //tell the remaining how big to make themselves
        hash_table_firstkey(sizes);
        char* sizes_key;
        long long sizes_cor;
        while (hash_table_nextkey(sizes, &sizes_key, (void**) &sizes_cor)) {
            hash_table_firstkey(comps);
            while (hash_table_nextkey(comps, &key, (void**) &value)) {
                if (strstr(key, sizes_key) != NULL) {
                    if (getenv("MPI_WORKER_CORES_PER") != NULL) { //check to see if we're passing this via env-var
                        sizes_cor = atoi(getenv("MPI_WORKER_CORES_PER"));
                    }
                    int sizes_cor_int = sizes_cor;
                    MPI_Send(&sizes_cor_int, 1, MPI_INT, value, 0, MPI_COMM_WORLD);
                }
            }
        }

        int cores = rank_0_cores;
        int cores_total = load_average_get_cpus();
        UINT64_T memtotal;
        UINT64_T memavail;
        host_memory_info_get(&memavail, &memtotal);
        int mem = ((memtotal / (1024 * 1024)) / cores_total) * cores; //MB

        char* sys_str = string_format("makeflow -T wq --port=%s -d all --local-cores=%i --local-memory=%i %s", port, 1, mem, makeflow_args);
        if (debug_base != NULL) {
            sys_str = string_format("makeflow -T wq --port=%s -dall --debug-file=%s.makeflow --local-cores=%i --local-memory=%i %s", port, debug_base, 1, mem, makeflow_args);
        }

        int k = 0;
        k = system(sys_str);
        fprintf(stderr, "Makeflow has finished! Sending kill to workers\n");
        hash_table_firstkey(comps);
        //char* key1;
        //long long value1;
        while (hash_table_nextkey(comps, &key, (void**) &value)) {
            fprintf(stderr, "Telling %lli to end worker\n", value);
            unsigned die = 10;
            MPI_Send(&die, 1, MPI_UNSIGNED, value, 0, MPI_COMM_WORLD);
        }

        if (cpout != NULL) {
            sys_str = string_format("cp -r `pwd`/* %s", cpout);
            system(sys_str);
        }

        hash_table_delete(comps);
        hash_table_delete(sizes);

        MPI_Finalize();
		fprintf(stderr,"Makeflow master thread, this is my k: %i\n",k);
        return k;

    } else { //we're a work_queue_worker

        fprintf(stderr, "Yay, i'm a worker starter, and my procname is: %s\n", procname);

        unsigned len = 0;
        MPI_Recv(&len, 1, MPI_UNSIGNED, 0, 0, MPI_COMM_WORLD, MPI_STATUS_IGNORE);
        char* master_ip = malloc(sizeof (char*)*len + 1);
        memset(master_ip, '\0', sizeof (char)*len + 1);
        MPI_Recv(master_ip, len, MPI_CHAR, 0, 0, MPI_COMM_WORLD, MPI_STATUS_IGNORE);

        fprintf(stderr, "Here is the master_ipaddr: %s\n", master_ip);

        int cores;
        MPI_Recv(&cores, 1, MPI_INT, 0, 0, MPI_COMM_WORLD, MPI_STATUS_IGNORE);

        //get a fair amount of memory, for now
        int cores_total = load_average_get_cpus();
        UINT64_T memtotal;
        UINT64_T memavail;
        host_memory_info_get(&memavail, &memtotal);
        fprintf(stderr, "worker: %i memtotal: %lu\n", mpi_rank, memtotal);
        int mem = ((memtotal / (1024 * 1024)) / cores_total) * cores; //gigabytes

        fprintf(stderr, "Calling printenv from worker: %i\n", mpi_rank);

        char* sys_str = string_format("work_queue_worker --timeout=86400 --cores=%i --memory=%i %s %s %s", cores, mem, master_ip, port, workqueue_args);
        if (debug_base != NULL) {
            sys_str = string_format("work_queue_worker --timeout=86400 -d all -o %s.workqueue.%i --cores=%i --memory=%i %s %s %s", debug_base, mpi_rank, cores, mem, master_ip, port, workqueue_args);
        }
        fprintf(stderr, "Rank %i: Starting Worker: %s\n", mpi_rank, sys_str);
        int pid = fork();
        int k = 0;
        if (pid == 0) {
            signal(SIGTERM, wq_handle);
            for (;;) {
                workqueue_pid = fork();
                if (workqueue_pid == 0) {
                    int argc;
                    char **argv;
                    string_split_quotes(sys_str, &argc, &argv);
                    execvp(argv[0], argv);
                    fprintf(stderr, "Oh no, workqueue execvp didn't work...");
					exit(1);
                }
                int status_wq;
                waitpid(workqueue_pid, &status_wq, 0);
            }
        } else {
            unsigned die;
            MPI_Recv(&die, 1, MPI_UNSIGNED, 0, 0, MPI_COMM_WORLD, MPI_STATUS_IGNORE);
            kill(pid, SIGTERM);
        }

        MPI_Finalize();
        return k;

    }

    return 0;
}
#else
#include <stdio.h>
int main(int argc, char** argv){
    fprintf(stdout,"To use this Program, please configure and compile cctools with MPI.\n");
}

#endif


/* vim: set noexpandtab tabstop=4: */
