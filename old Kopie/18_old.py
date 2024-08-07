# Copyright 2017 Google Inc.
from google.cloud.security import FORSETI_SECURITY_HOME_ENV_VAR
CONFIGS_FILE = os.path.abspath(
    os.path.join(os.environ.get(FORSETI_SECURITY_HOME_ENV_VAR),
                 'config', 'db.yaml'))
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Provides the database connector."""

import os
import gflags as flags

import MySQLdb
from MySQLdb import OperationalError
import yaml

from google.cloud.security.common.data_access.errors import MySQLError
from google.cloud.security.common.util.log_util import LogUtil

LOGGER = LogUtil.setup_logging(__name__)

FLAGS = flags.FLAGS
flags.DEFINE_string('db_host', None, 'Database hostname/IP address')
flags.DEFINE_string('db_name', None, 'Database name')
flags.DEFINE_string('db_user', None, 'Database user')
flags.DEFINE_string('db_passwd', None, 'Database password')
flags.mark_flag_as_required('db_host')
flags.mark_flag_as_required('db_name')
flags.mark_flag_as_required('db_user')
flags.mark_flag_as_required('db_passwd')


class _DbConnector(object):
    """Database connector."""

    def __init__(self):
        """Initialize the db connector.

        Raises:
            MySQLError: An error with MySQL has occurred.
        """
        configs = FLAGS.FlagValuesDict()

        try:
            self.conn = MySQLdb.connect(
                host=configs['db_host'],
                user=configs['db_user'],
                passwd=configs['db_passwd'],
                db=configs['db_name'],
                local_infile=1)
        except OperationalError as e:
            LOGGER.error('Unable to create mysql connector:\n{0}'.format(e))
            raise MySQLError('DB Connector', e)

    def __del__(self):
        """Closes the database connection."""
        try:
            self.conn.close()
        except AttributeError:
            # This happens when conn is not created in the first place,
            # thus does not need to be cleaned up.
            pass
