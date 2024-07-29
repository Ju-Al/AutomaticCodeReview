			resp.sendError(HttpServletResponse.SC_SERVICE_UNAVAILABLE,
					"In the node initialization, unable to process any requests at this time");
			return;
		}

		try {
			// If an unrecoverable exception occurs on this node, the write request operation shall not be processed
			// This is a very important warning message !!!
			if (downgrading) {
				resp.sendError(HttpServletResponse.SC_SERVICE_UNAVAILABLE,
						"Unable to process the request at this time: System triggered degradation");
				return;
			}

			chain.doFilter(req, response);
		}
		catch (AccessControlException e) {
			resp.sendError(HttpServletResponse.SC_FORBIDDEN,
					"access denied: " + ExceptionUtil.getAllExceptionMsg(e));
			return;
		}
		catch (Throwable e) {
			resp.sendError(HttpServletResponse.SC_INTERNAL_SERVER_ERROR,
					"Server failed," + e.toString());
			return;
		}
	}

	@Override
	public void destroy() {

	}

	private void listenerSelfInCluster() {
		protocol.protocolMetaData().subscribe(Constants.CONFIG_MODEL_RAFT_GROUP,
				com.alibaba.nacos.consistency.cp.Constants.RAFT_GROUP_MEMBER,
				new Observer() {
					@Override
					public void update(Observable o, Object arg) {
						final List<String> peers = (List<String>) arg;
						final Member self = memberManager.getSelf();
						final String raftAddress = self.getIp() + ":" + self
								.getExtendVal(MemberMetaDataConstants.RAFT_PORT);
						// Only when you are in the cluster and the current Leader is
						// elected can you provide external services
						openService = peers.contains(raftAddress);
					}
				});
	}

	private void registerSubscribe() {
		NotifyCenter.registerSubscribe(new SmartSubscribe() {

			@Override
			public void onEvent(Event event) {
				// @JustForTest
				// This event only happens in the case of unit tests
				if (event instanceof RaftDBErrorRecoverEvent) {
					downgrading = false;
					return;
				}
				if (event instanceof RaftDBErrorEvent) {
					downgrading = true;
				}
			}

			@Override
			public boolean canNotify(Event event) {
				return (event instanceof RaftDBErrorEvent)
						|| (event instanceof RaftDBErrorRecoverEvent);
			}
		});
	}

}
