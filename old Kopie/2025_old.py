# Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.
sample_size = 100
premade_batch = [np.random.rand(sample_size)]


def isclose(a, b, rel_tol=1e-09, abs_tol=0.0):
    """
    Python2.7 doesn't have math.isclose()
    """
    return abs(a - b) <= max(rel_tol * max(abs(a), abs(b)), abs_tol)


def preemph(signal, coeff=0.0):
    ret = np.copy(signal)
    if coeff == 0.0:
        return ret
    for i in range(sample_size - 1, 0, -1):
        ret[i] -= coeff * ret[i - 1]
    ret[0] -= coeff * ret[0]
    return ret
    def __init__(self, preemph_coeff=0., num_threads=1):
        super(PreemphasisPipeline, self).__init__(1, num_threads, 0)
        self.preemph = ops.PreemphasisFilter(preemph_coeff=preemph_coeff, device="cpu")
        return self.preemph(self.data)
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

from __future__ import print_function
from nvidia.dali.pipeline import Pipeline
import nvidia.dali.ops as ops
import nvidia.dali.types as types
import numpy as np
from functools import partial
from test_utils import compare_pipelines
from test_utils import RandomlyShapedDataIterator

SEED = 12345

def preemph_func(coeff, signal):
    in_shape = signal.shape
    assert(len(in_shape) == 1)  # 1D
    out = np.copy(signal)
    out[0]  -= coeff * signal[0]
    out[1:] -= coeff * signal[0:in_shape[0]-1]
    return out

class PreemphasisPipeline(Pipeline):
    def __init__(self, device, batch_size, iterator, preemph_coeff=0.97, per_sample_coeff=False, num_threads=4, device_id=0):
        super(PreemphasisPipeline, self).__init__(batch_size, num_threads, device_id, seed=SEED)
        self.device = device
        self.iterator = iterator
        self.ext_src = ops.ExternalSource()
        self.per_sample_coeff = per_sample_coeff
        self.uniform = ops.Uniform(range=(0.5, 0.97), seed=1234)
        if self.per_sample_coeff:
            self.preemph = ops.PreemphasisFilter(device=device)
        else:
            self.preemph = ops.PreemphasisFilter(device=device, preemph_coeff=preemph_coeff)

    def define_graph(self):
        self.data = self.ext_src()
        out = self.data.gpu() if self.device == 'gpu' else self.data
        if self.per_sample_coeff:
            preemph_coeff_arg = self.uniform()
            return self.preemph(out, preemph_coeff=preemph_coeff_arg)
        else:
            return self.preemph(out)

    def iter_setup(self):
        data = self.iterator.next()
        self.feed_input(self.data, data)

class PreemphasisPythonPipeline(Pipeline):
    def __init__(self, device, batch_size, iterator, preemph_coeff=0.97, per_sample_coeff=False, num_threads=4, device_id=0):
        super(PreemphasisPythonPipeline, self).__init__(batch_size, num_threads, device_id, seed=SEED,
                                                        exec_async=False, exec_pipelined=False)
        self.device = "cpu"
        self.iterator = iterator
        self.ext_src = ops.ExternalSource()
        self.per_sample_coeff = per_sample_coeff
        self.uniform = ops.Uniform(range=(0.5, 0.97), seed=1234)
        if self.per_sample_coeff:
            function = preemph_func
        else:
            function = partial(preemph_func, preemph_coeff)
        self.preemph = ops.PythonFunction(function=function)

    def define_graph(self):
        self.data = self.ext_src()
        if self.per_sample_coeff:
            coef = self.uniform()
            return self.preemph(coef, self.data)
        else:
            return self.preemph(self.data)

    def iter_setup(self):
        data = self.iterator.next()
        self.feed_input(self.data, data)

def check_preemphasis_operator(device, batch_size, preemph_coeff, per_sample_coeff):
    eii1 = RandomlyShapedDataIterator(batch_size, min_shape=(100, ), max_shape=(10000, ), dtype=np.float32)
    eii2 = RandomlyShapedDataIterator(batch_size, min_shape=(100, ), max_shape=(10000, ), dtype=np.float32)
    compare_pipelines(
        PreemphasisPipeline(device, batch_size, iter(eii1), preemph_coeff=preemph_coeff, per_sample_coeff=per_sample_coeff),
        PreemphasisPythonPipeline(device, batch_size, iter(eii2), preemph_coeff=preemph_coeff, per_sample_coeff=per_sample_coeff),
        batch_size=batch_size, N_iterations=3)

def test_preemphasis_operator():
    for device in ['cpu', 'gpu']:
        for batch_size in [1, 3, 128]:
            for coef, per_sample_coeff in [(0.97, False), (0.0, False), (None, True)]:
                yield check_preemphasis_operator, device, batch_size, coef, per_sample_coeff
