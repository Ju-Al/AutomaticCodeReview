"""Farthest Point Sampler for mxnet Geometry package"""

from mxnet import nd
from mxnet.gluon import nn
import numpy as np

from ..capi import farthest_point_sampler

class FarthestPointSampler(nn.Block):
    """Farthest Point Sampler

    Parameters
    ----------
    npoints : int
        The number of points to sample in each batch.
    """
    def __init__(self, npoints):
        super(FarthestPointSampler, self).__init__()
        self.npoints = npoints

    def forward(self, pos):
        r"""Memory allocation and sampling

        Parameters
        ----------
        pos : tensor
            The positional tensor of shape (B, N, C)

        Returns
        -------
        tensor of shape (B, self.npoints)
            The sampled indices in each batch.
        """
        ctx = pos.context
        B, N, C = pos.shape
        pos = pos.reshape(-1, C)
        dist = nd.zeros((B * N), dtype=pos.dtype, ctx=ctx)
        start_idx = nd.random.randint(0, N - 1, (B, ), dtype=np.int, ctx=ctx)
        result = nd.zeros((self.npoints * B), dtype=np.int, ctx=ctx)
        farthest_point_sampler(pos, B, self.npoints, dist, start_idx, result)
        return result.reshape(B, self.npoints)
