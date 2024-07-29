    except JSONDecodeError:
      raise errors.UnableToParseFile('could not decode json.')

    try:
      if 'logName' in file_json[0]:
        self._ParseLogJSON(parser_mediator, file_json)
    except ValueError as exception:
      if exception == 'No JSON object could be decoded':
        raise errors.UnableToParseFile(exception)
      raise


manager.ParsersManager.RegisterParser(GCPLogsParser)
