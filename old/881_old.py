import os
class _ScriptModificationHandler(PatternMatchingEventHandler):
    def __init__(self, callback):
        # We could enumerate all relevant *.py files (as werkzeug does it),
        # but our case looks like it isn't as simple as enumerating sys.modules.
        # This should be good enough for now.
            patterns=["*.py"]
import sys
from watchdog.events import RegexMatchingEventHandler 
if sys.platform == 'darwin':
    from watchdog.observers.polling import PollingObserver as Observer
else:
    from watchdog.observers import Observer
# The OSX reloader in watchdog 0.8.3 breaks when unobserving paths. 
# We use the PollingObserver instead.

_observers = {}


def watch(script, callback):
    if script in _observers:
        raise RuntimeError("Script already observed")
    script_dir = os.path.dirname(os.path.abspath(script.args[0]))
    script_name = os.path.basename(script.args[0])
    event_handler = _ScriptModificationHandler(callback, filename=script_name)
    observer = Observer()
    observer.schedule(event_handler, script_dir)
    observer.start()
    _observers[script] = observer


def unwatch(script):
    observer = _observers.pop(script, None)
    if observer:
        observer.stop()
        observer.join()


class _ScriptModificationHandler(RegexMatchingEventHandler):
    def __init__(self, callback, filename='.*'):

        super(_ScriptModificationHandler, self).__init__(
            ignore_directories=True,
            regexes=['.*'+filename]
        )
        self.callback = callback
        self.filename = filename 

    def on_modified(self, event):
        self.callback()

__all__ = ["watch", "unwatch"]

