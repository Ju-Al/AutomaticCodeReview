
	struct state state = {
		.angle = 0.0,
		.output_add = { .notify = output_add },
		.output_remove = { .notify = output_remove }
	};

	wl_list_init(&state.outputs);
	wl_list_init(&state.config);
	wl_list_init(&state.output_add.link);
	wl_list_init(&state.output_remove.link);
	clock_gettime(CLOCK_MONOTONIC, &state.last_frame);

	parse_args(argc, argv, &state.config);

	struct wl_display *display = wl_display_create();
	struct wl_event_loop *event_loop = wl_display_get_event_loop(display);

	struct wlr_session *session = wlr_session_start(display);
	if (!session) {
		return 1;
	}

	struct wlr_backend *wlr = wlr_backend_autocreate(display, session);
	if (!wlr) {
		return 1;
	}

	wl_signal_add(&wlr->events.output_add, &state.output_add);
	wl_signal_add(&wlr->events.output_remove, &state.output_remove);

	if (!wlr_backend_init(wlr)) {
		return 1;
	}

	init_gl(&state.gl);

	bool done = false;
	struct wl_event_source *timer = wl_event_loop_add_timer(event_loop,
		timer_done, &done);

	wl_event_source_timer_update(timer, 10000);

	while (!done) {
		wl_event_loop_dispatch(event_loop, 0);
	}

	cleanup_gl(&state.gl);

	wl_event_source_remove(timer);
	wlr_backend_destroy(wlr);
	wlr_session_finish(session);
	wl_display_destroy(display);

	struct output_config *ptr, *tmp;
	wl_list_for_each_safe(ptr, tmp, &state.config, link) {
		free(ptr);
	}
}
