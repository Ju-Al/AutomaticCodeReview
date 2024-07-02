"""Copyright 2008 Orbitz WorldWide

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License."""

import imp
import os
import socket
import time
import sys
import calendar
import pytz
from datetime import datetime
from os.path import splitext, basename, relpath
from shutil import move
from tempfile import mkstemp
try:
  import cPickle as pickle
  USING_CPICKLE = True
except:
  import pickle
  USING_CPICKLE = False

try:
  from cStringIO import StringIO
except ImportError:
  from StringIO import StringIO

from django.conf import settings
from django.utils.timezone import make_aware
from graphite.logger import log


# There are a couple different json modules floating around out there with
# different APIs. Hide the ugliness here.
try:
  import json
except ImportError:
  import simplejson as json

if hasattr(json, 'read') and not hasattr(json, 'loads'):
  json.loads = json.read
  json.dumps = json.write
  json.load = lambda file: json.read( file.read() )
  json.dump = lambda obj, file: file.write( json.write(obj) )

def epoch(dt):
  """
  Returns the epoch timestamp of a timezone-aware datetime object.
  """
  if not dt.tzinfo:
    return calendar.timegm(make_aware(dt, pytz.timezone(settings.TIME_ZONE)).astimezone(pytz.utc).timetuple())
  return calendar.timegm(dt.astimezone(pytz.utc).timetuple())


def epoch_to_dt(timestamp):
    """
    Returns the timezone-aware datetime of an epoch timestamp.
    """
    return make_aware(datetime.utcfromtimestamp(timestamp), pytz.utc)

def timebounds(requestContext):
  startTime = int(epoch(requestContext['startTime']))
  endTime = int(epoch(requestContext['endTime']))
  now = int(epoch(requestContext['now']))

  return (startTime, endTime, now)

def is_local_interface(host):
  is_ipv6 = False
  if ':' not in host:
    pass
  elif host.count(':') == 1:
    host = host.split(':', 1)[0]
  else:
    is_ipv6 = True

    if host.find('[', 0, 2) != -1:
      last_bracket_position  = host.rfind(']')
      last_colon_position = host.rfind(':')
      if last_colon_position > last_bracket_position:
        host = host.rsplit(':', 1)[0]
      host = host.strip('[]')

  try:
    if is_ipv6:
      sock = socket.socket(socket.AF_INET6, socket.SOCK_DGRAM)
    else:
      sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind( (host, 0) )
  except:
    return False
  finally:
    sock.close()

  return True


def is_pattern(s):
   return '*' in s or '?' in s or '[' in s or '{' in s

def is_escaped_pattern(s):
  for symbol in '*?[{':
    i = s.find(symbol)
    if i > 0:
      if s[i-1] == '\\':
        return True
  return False

def find_escaped_pattern_fields(pattern_string):
  pattern_parts = pattern_string.split('.')
  for index,part in enumerate(pattern_parts):
    if is_escaped_pattern(part):
      yield index


def load_module(module_path, member=None):
  module_name = splitext(basename(module_path))[0]
  module_file = open(module_path, 'U')
  description = ('.py', 'U', imp.PY_SOURCE)
  module = imp.load_module(module_name, module_file, module_path, description)
  if member:
    return getattr(module, member)
  else:
    return module

def timestamp(dt):
  "Convert a datetime object into epoch time"
  return time.mktime(dt.timetuple())

def deltaseconds(timedelta):
  "Convert a timedelta object into seconds (same as timedelta.total_seconds() in Python 2.7+)"
  return (timedelta.microseconds + (timedelta.seconds + timedelta.days * 24 * 3600) * 10**6) / 10**6


# This whole song & dance is due to pickle being insecure
# The SafeUnpickler classes were largely derived from
# http://nadiana.com/python-pickle-insecure
# This code also lives in carbon.util
if USING_CPICKLE:
  class SafeUnpickler(object):
    PICKLE_SAFE = {
      'copy_reg': set(['_reconstructor']),
      '__builtin__': set(['object', 'list', 'set']),
      'collections': set(['deque']),
      'graphite.render.datalib': set(['TimeSeries']),
      'graphite.intervals': set(['Interval', 'IntervalSet']),
    }

    @classmethod
    def find_class(cls, module, name):
      if not module in cls.PICKLE_SAFE:
        raise pickle.UnpicklingError('Attempting to unpickle unsafe module %s' % module)
      __import__(module)
      mod = sys.modules[module]
      if not name in cls.PICKLE_SAFE[module]:
        raise pickle.UnpicklingError('Attempting to unpickle unsafe class %s' % name)
      return getattr(mod, name)

    @classmethod
    def loads(cls, pickle_string):
      pickle_obj = pickle.Unpickler(StringIO(pickle_string))
      pickle_obj.find_global = cls.find_class
      return pickle_obj.load()

else:
  class SafeUnpickler(pickle.Unpickler):
    PICKLE_SAFE = {
      'copy_reg': set(['_reconstructor']),
      '__builtin__': set(['object', 'list', 'set']),
      'collections': set(['deque']),
      'graphite.render.datalib': set(['TimeSeries']),
      'graphite.intervals': set(['Interval', 'IntervalSet']),
    }

    def find_class(self, module, name):
      if not module in self.PICKLE_SAFE:
        raise pickle.UnpicklingError('Attempting to unpickle unsafe module %s' % module)
      __import__(module)
      mod = sys.modules[module]
      if not name in self.PICKLE_SAFE[module]:
        raise pickle.UnpicklingError('Attempting to unpickle unsafe class %s' % name)
      return getattr(mod, name)

    @classmethod
    def loads(cls, pickle_string):
      return cls(StringIO(pickle_string)).load()

unpickle = SafeUnpickler


def write_index(whisper_dir=None, ceres_dir=None, index=None):
  if not whisper_dir:
    whisper_dir = settings.WHISPER_DIR
  if not ceres_dir:
    ceres_dir = settings.CERES_DIR
  if not index:
    index = settings.INDEX_FILE
  try:
    fd, tmp = mkstemp()
    try:
      tmp_index = os.fdopen(fd, 'wt')
      build_index(whisper_dir, ".wsp", tmp_index)
      build_index(ceres_dir, ".ceres-node", tmp_index)
    finally:
      tmp_index.close()
    move(tmp, index)
  finally:
    try:
      os.unlink(tmp)
    except:
      pass
  return None


def logtime(custom_msg=False, custom_name=False):
  def wrap(f):
    def wrapped_f(*args, **kwargs):
      msg = 'completed in'
      name = f.__module__ + '.' + f.__name__

      t = time.time()
      if custom_msg:
        def set_msg(msg):
          wrapped_f.msg = msg

        kwargs['msg_setter'] = set_msg
      if custom_name:
        def set_name(name):
          wrapped_f.name = name

        kwargs['name_setter'] = set_name

      try:
        res = f(*args, **kwargs)
      except:
        msg = 'failed in'
        raise
      finally:
        msg = getattr(wrapped_f, 'msg', msg)
        name = getattr(wrapped_f, 'name', name)

        log.info(
          '{name} :: {msg} {sec:.6}s'.format(
            name=name,
            msg=msg,
            sec=time.time() - t,
          )
        )

      return res
    return wrapped_f
  return wrap


@logtime(custom_msg=True)
def build_index(base_path, extension, fd, msg_setter=None):
  t = time.time()
  total_entries = 0
  contents = os.walk(base_path, followlinks=True)
  extension_len = len(extension)
  for (dirpath, dirnames, filenames) in contents:
    path = relpath(dirpath, base_path).replace('/', '.')
    for metric in filenames:
      if metric.endswith(extension):
        metric = metric[:-extension_len]
      else:
        continue
      line = "{0}.{1}\n".format(path, metric)
      total_entries += 1
      fd.write(line)
  fd.flush()
  msg_setter("[IndexSearcher] index rebuild of \"%s\" took" % base_path)
  return None