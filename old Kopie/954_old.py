import traceback

class LogExceptionsMiddleware(object):
    def process_exception(self, request, exception):
        log.exception(traceback.format_exc())
        return None
