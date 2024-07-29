      LOG.error(e);
      return new JsonRpcErrorResponse(request.getId(), JsonRpcError.FIND_PRIVACY_GROUP_ERROR);
    }
  }
}
