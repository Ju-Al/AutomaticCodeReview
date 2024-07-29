		switch(c) {
		case 'm':
			pfs_mountfile_parse_file(optarg);
			break;
		case 'M':
			pfs_mountfile_parse_string(optarg);
			break;
		case 'h':
		case 'v':
		case 'l':
		case LONG_OPT_PARROT_PATH:
			break;
		default:
			show_help();
			return 1;
		}
	}

	if(optind >= argc) {
		show_help();
		return 1;
	}

	if (execvp(argv[optind], &argv[optind]) < 0) {
		fatal("failed to exec %s: %s\n", argv[optind], strerror(errno));
	}
}

/* vim: set noexpandtab tabstop=4: */
