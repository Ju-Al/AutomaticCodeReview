
    /* start to save state */
    for (i = 0; i < inbufcnt; ++i) {
        self->state.bufs.entries[i] = inbufs[i];
    }
    self->state.bufs.size = inbufcnt;
    self->state.is_final = is_final;

    /* if there's token, we try to send */
    if (self->tokens > 0)
        real_send(self);
}

static void on_stop(h2o_ostream_t *_self, h2o_req_t *req) {
    throttle_resp_t *self = (void *)_self;
    if (h2o_timeout_is_linked(&self->timeout_entry)) {
        h2o_timeout_unlink(&self->timeout_entry);
    }
}

static void on_setup_ostream(h2o_filter_t *self, h2o_req_t *req, h2o_ostream_t **slot) {
    throttle_resp_t *throttle;
    h2o_iovec_t traffic_header_value;
    size_t traffic_limit;

    if (req->res.status != 200)
        goto Next;
    if (h2o_memis(req->input.method.base, req->input.method.len, H2O_STRLIT("HEAD")))
        goto Next;

    ssize_t xt_index;
    if ((xt_index = h2o_find_header(&req->res.headers, H2O_TOKEN_X_TRAFFIC, -1)) == -1)
        goto Next;

    traffic_header_value = req->res.headers.entries[xt_index].value;
    char *buf = traffic_header_value.base;

    if (H2O_UNLIKELY((traffic_limit = h2o_strtosizefwd(&buf, traffic_header_value.len)) == SIZE_MAX))
        goto Next;

    throttle = (void *)h2o_add_ostream(req, sizeof(throttle_resp_t), slot);

    /* calculate the token increment per 100ms */
    throttle->token_inc = traffic_limit * HUNDRED_MS / ONE_SECOND;
    if (req->preferred_chunk_size > throttle->token_inc)
        req->preferred_chunk_size = throttle->token_inc;

    h2o_delete_header(&req->res.headers, xt_index);

    throttle->super.do_send = on_send;
    throttle->super.stop = on_stop;
    throttle->ctx = req->conn->ctx;
    throttle->req = req;
    throttle->state.bufs.capacity = 0;
    throttle->state.bufs.size = 0;
    throttle->timeout_entry = (h2o_timeout_entry_t){};
    throttle->timeout_entry.cb = add_token;
    throttle->tokens = throttle->token_inc;
    slot = &throttle->super.next;

    h2o_timeout_link(throttle->ctx->loop, &throttle->ctx->hundred_ms_timeout, &throttle->timeout_entry);

  Next:
    h2o_setup_next_ostream(req, slot);
}

void h2o_throttle_resp_register(h2o_pathconf_t *pathconf) {
    h2o_filter_t *self = h2o_create_filter(pathconf, sizeof(*self));
    self->on_setup_ostream = on_setup_ostream;
}
