# Copyright 2019 The TensorFlow Authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================
"""Utilities similar to tf.python.platform.resource_loader."""

from distutils.version import LooseVersion
import os
import warnings

import tensorflow as tf

MIN_TF_VERSION_FOR_ABI_COMPATIBILITY = "2.1.0"
MAX_TF_VERSION_FOR_ABI_COMPATIBILITY = "2.2.0"
abi_warning_already_raised = False


def get_project_root():
    """Returns project root folder."""
    return os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def get_path_to_datafile(path):
    """Get the path to the specified file in the data dependencies.

    The path is relative to tensorflow_addons/

    Args:
      path: a string resource path relative to tensorflow_addons/
    Returns:
      The path to the specified data file
    """
    root_dir = get_project_root()
    return os.path.join(root_dir, path.replace("/", os.sep))


class LazySO:
    def __init__(self, relative_path):
        self.relative_path = relative_path
        self._ops = None

    @property
    def ops(self):
        if self._ops is None:
            self._ops = tf.load_op_library(get_path_to_datafile(self.relative_path))
        return self._ops
            "below {}.\n Note that nightly versions of TensorFlow are not supported."
            "\n\n"
            "The last solution is to find the TensorFlow Addons version that has "
            "custom ops compatible with the TensorFlow installed on your "
            "system. To do that, refer to the readme: "
            "https://github.com/tensorflow/addons"
            "".format(
                tf.__version__,
                self.relative_path,
                MIN_TF_VERSION_FOR_ABI_COMPATIBILITY,
                MIN_TF_VERSION_FOR_ABI_COMPATIBILITY,
                MAX_TF_VERSION_FOR_ABI_COMPATIBILITY,
            ),
            UserWarning,
        )
        abi_warning_already_raised = True


def abi_is_compatible():
    if "dev" in tf.__version__:
        # tf-nightly
        return False

    min_version = LooseVersion(MIN_TF_VERSION_FOR_ABI_COMPATIBILITY)
    max_version = LooseVersion(MAX_TF_VERSION_FOR_ABI_COMPATIBILITY)
    return min_version <= LooseVersion(tf.__version__) < max_version