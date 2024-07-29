      Configurator.setAllLevels("", logLevel);
      return new JsonRpcSuccessResponse(request.getId());
    } catch (InvalidJsonRpcParameters invalidJsonRpcParameters) {
      return new JsonRpcErrorResponse(request.getId(), JsonRpcError.INVALID_PARAMS);
    }
  }
}
