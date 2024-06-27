"""Module for graph transformation utilities."""
    'as_heterograph']

from collections.abc import Iterable, Mapping
from collections import defaultdict
import numpy as np
from scipy import sparse
from ._ffi.function import _init_api
from .graph import DGLGraph
from .heterograph import DGLHeteroGraph
from . import ndarray as nd
from . import backend as F
from .graph_index import from_coo
from .graph_index import _get_halo_subgraph_inner_node
from .graph import unbatch
from .convert import graph, bipartite
from . import utils
from .base import EID, NID
from . import ndarray as nd
from .frame import Frame, FrameRef


__all__ = [
    'line_graph',
    'khop_adj',
    'khop_graph',
    'reverse',
    'to_simple_graph',
    'to_bidirected',
    'laplacian_lambda_max',
    'knn_graph',
    'segmented_knn_graph',
    'add_self_loop',
    'remove_self_loop',
    'metapath_reachable_graph',
    'compact_graphs',
    'to_block',
    'to_simple',
    'in_subgraph',
    'out_subgraph',
    'remove_edges',
    'as_immutable_graph',
    'as_heterograph',
    'sort_csr_',
    'sort_csc_',
    'sort_csr',
    'sort_csc']


def pairwise_squared_distance(x):
    """
    x : (n_samples, n_points, dims)
    return : (n_samples, n_points, n_points)
    """
    x2s = F.sum(x * x, -1, True)
    # assuming that __matmul__ is always implemented (true for PyTorch, MXNet and Chainer)
    return x2s + F.swapaxes(x2s, -1, -2) - 2 * x @ F.swapaxes(x, -1, -2)

#pylint: disable=invalid-name
def knn_graph(x, k):
    """Transforms the given point set to a directed graph, whose coordinates
    are given as a matrix. The predecessors of each point are its k-nearest
    neighbors.

    If a 3D tensor is given instead, then each row would be transformed into
    a separate graph.  The graphs will be unioned.

    Parameters
    ----------
    x : Tensor
        The input tensor.

        If 2D, each row of ``x`` corresponds to a node.

        If 3D, a k-NN graph would be constructed for each row.  Then
        the graphs are unioned.
    k : int
        The number of neighbors

    Returns
    -------
    DGLGraph
        The graph.  The node IDs are in the same order as ``x``.
    """
    if F.ndim(x) == 2:
        x = F.unsqueeze(x, 0)
    n_samples, n_points, _ = F.shape(x)

    dist = pairwise_squared_distance(x)
    k_indices = F.argtopk(dist, k, 2, descending=False)
    dst = F.copy_to(k_indices, F.cpu())

    src = F.zeros_like(dst) + F.reshape(F.arange(0, n_points), (1, -1, 1))

    per_sample_offset = F.reshape(F.arange(0, n_samples) * n_points, (-1, 1, 1))
    dst += per_sample_offset
    src += per_sample_offset
    dst = F.reshape(dst, (-1,))
    src = F.reshape(src, (-1,))
    adj = sparse.csr_matrix((F.asnumpy(F.zeros_like(dst) + 1), (F.asnumpy(dst), F.asnumpy(src))))

    g = DGLGraph(adj, readonly=True)
    return g

#pylint: disable=invalid-name
def segmented_knn_graph(x, k, segs):
    """Transforms the given point set to a directed graph, whose coordinates
    are given as a matrix.  The predecessors of each point are its k-nearest
    neighbors.

    The matrices are concatenated along the first axis, and are segmented by
    ``segs``.  Each block would be transformed into a separate graph.  The
    graphs will be unioned.

    Parameters
    ----------
    x : Tensor
        The input tensor.
    k : int
        The number of neighbors
    segs : iterable of int
        Number of points of each point set.
        Must sum up to the number of rows in ``x``.

    Returns
    -------
    DGLGraph
        The graph.  The node IDs are in the same order as ``x``.
    """
    n_total_points, _ = F.shape(x)
    offset = np.insert(np.cumsum(segs), 0, 0)

    h_list = F.split(x, segs, 0)
    dst = [
        F.argtopk(pairwise_squared_distance(h_g), k, 1, descending=False) +
        offset[i]
        for i, h_g in enumerate(h_list)]
    dst = F.cat(dst, 0)
    src = F.arange(0, n_total_points).unsqueeze(1).expand(n_total_points, k)

    dst = F.reshape(dst, (-1,))
    src = F.reshape(src, (-1,))
    adj = sparse.csr_matrix((F.asnumpy(F.zeros_like(dst) + 1), (F.asnumpy(dst), F.asnumpy(src))))

    g = DGLGraph(adj, readonly=True)
    return g

def line_graph(g, backtracking=True, shared=False):
    """Return the line graph of this graph.

    Parameters
    ----------
    g : dgl.DGLGraph
        The input graph.
    backtracking : bool, optional
        Whether the returned line graph is backtracking.
    shared : bool, optional
        Whether the returned line graph shares representations with `self`.

    Returns
    -------
    DGLGraph
        The line graph of this graph.
    """
    graph_data = g._graph.line_graph(backtracking)
    node_frame = g._edge_frame if shared else None
    return DGLGraph(graph_data, node_frame)

def khop_adj(g, k):
    """Return the matrix of :math:`A^k` where :math:`A` is the adjacency matrix of :math:`g`,
    where a row represents the destination and a column represents the source.

    Parameters
    ----------
    g : dgl.DGLGraph
        The input graph.
    k : int
        The :math:`k` in :math:`A^k`.

    Returns
    -------
    tensor
        The returned tensor, dtype is ``np.float32``.

    Examples
    --------

    >>> import dgl
    >>> g = dgl.DGLGraph()
    >>> g.add_nodes(5)
    >>> g.add_edges([0,1,2,3,4,0,1,2,3,4], [0,1,2,3,4,1,2,3,4,0])
    >>> dgl.khop_adj(g, 1)
    tensor([[1., 0., 0., 0., 1.],
            [1., 1., 0., 0., 0.],
            [0., 1., 1., 0., 0.],
            [0., 0., 1., 1., 0.],
            [0., 0., 0., 1., 1.]])
    >>> dgl.khop_adj(g, 3)
    tensor([[1., 0., 1., 3., 3.],
            [3., 1., 0., 1., 3.],
            [3., 3., 1., 0., 1.],
            [1., 3., 3., 1., 0.],
            [0., 1., 3., 3., 1.]])
    """
    adj_k = g.adjacency_matrix_scipy(return_edge_ids=False) ** k
    return F.tensor(adj_k.todense().astype(np.float32))

def khop_graph(g, k):
    """Return the graph that includes all :math:`k`-hop neighbors of the given graph as edges.
    The adjacency matrix of the returned graph is :math:`A^k`
    (where :math:`A` is the adjacency matrix of :math:`g`).

    Parameters
    ----------
    g : dgl.DGLGraph
        The input graph.
    k : int
        The :math:`k` in `k`-hop graph.

    Returns
    -------
    dgl.DGLGraph
        The returned ``DGLGraph``.

    Examples
    --------

    Below gives an easy example:

    >>> import dgl
    >>> g = dgl.DGLGraph()
    >>> g.add_nodes(3)
    >>> g.add_edges([0, 1], [1, 2])
    >>> g_2 = dgl.transform.khop_graph(g, 2)
    >>> print(g_2.edges())
    (tensor([0]), tensor([2]))

    A more complicated example:

    >>> import dgl
    >>> g = dgl.DGLGraph()
    >>> g.add_nodes(5)
    >>> g.add_edges([0,1,2,3,4,0,1,2,3,4], [0,1,2,3,4,1,2,3,4,0])
    >>> dgl.khop_graph(g, 1)
    DGLGraph(num_nodes=5, num_edges=10,
             ndata_schemes={}
             edata_schemes={})
    >>> dgl.khop_graph(g, 3)
    DGLGraph(num_nodes=5, num_edges=40,
             ndata_schemes={}
             edata_schemes={})
    """
    n = g.number_of_nodes()
    adj_k = g.adjacency_matrix_scipy(transpose=True, return_edge_ids=False) ** k
    adj_k = adj_k.tocoo()
    multiplicity = adj_k.data
    row = np.repeat(adj_k.row, multiplicity)
    col = np.repeat(adj_k.col, multiplicity)
    # TODO(zihao): we should support creating multi-graph from scipy sparse matrix
    # in the future.
    return DGLGraph(from_coo(n, row, col, True))

def reverse(g, share_ndata=False, share_edata=False):
    """Return the reverse of a graph

    The reverse (also called converse, transpose) of a directed graph is another directed
    graph on the same nodes with edges reversed in terms of direction.

    Given a :class:`DGLGraph` object, we return another :class:`DGLGraph` object
    representing its reverse.

    Notes
    -----
    * We do not dynamically update the topology of a graph once that of its reverse changes.
      This can be particularly problematic when the node/edge attrs are shared. For example,
      if the topology of both the original graph and its reverse get changed independently,
      you can get a mismatched node/edge feature.

    Parameters
    ----------
    g : dgl.DGLGraph
        The input graph.
    share_ndata: bool, optional
        If True, the original graph and the reversed graph share memory for node attributes.
        Otherwise the reversed graph will not be initialized with node attributes.
    share_edata: bool, optional
        If True, the original graph and the reversed graph share memory for edge attributes.
        Otherwise the reversed graph will not have edge attributes.

    Examples
    --------
    Create a graph to reverse.

    >>> import dgl
    >>> import torch as th
    >>> g = dgl.DGLGraph()
    >>> g.add_nodes(3)
    >>> g.add_edges([0, 1, 2], [1, 2, 0])
    >>> g.ndata['h'] = th.tensor([[0.], [1.], [2.]])
    >>> g.edata['h'] = th.tensor([[3.], [4.], [5.]])

    Reverse the graph and examine its structure.

    >>> rg = g.reverse(share_ndata=True, share_edata=True)
    >>> print(rg)
    DGLGraph with 3 nodes and 3 edges.
    Node data: {'h': Scheme(shape=(1,), dtype=torch.float32)}
    Edge data: {'h': Scheme(shape=(1,), dtype=torch.float32)}

    The edges are reversed now.

    >>> rg.has_edges_between([1, 2, 0], [0, 1, 2])
    tensor([1, 1, 1])

    Reversed edges have the same feature as the original ones.

    >>> g.edges[[0, 2], [1, 0]].data['h'] == rg.edges[[1, 0], [0, 2]].data['h']
    tensor([[1],
            [1]], dtype=torch.uint8)

    The node/edge features of the reversed graph share memory with the original
    graph, which is helpful for both forward computation and back propagation.

    >>> g.ndata['h'] = g.ndata['h'] + 1
    >>> rg.ndata['h']
    tensor([[1.],
            [2.],
            [3.]])
    """
    g_reversed = DGLGraph()
    g_reversed.add_nodes(g.number_of_nodes())
    g_edges = g.all_edges(order='eid')
    g_reversed.add_edges(g_edges[1], g_edges[0])
    g_reversed._batch_num_nodes = g._batch_num_nodes
    g_reversed._batch_num_edges = g._batch_num_edges
    if share_ndata:
        g_reversed._node_frame = g._node_frame
    if share_edata:
        g_reversed._edge_frame = g._edge_frame
    return g_reversed

def to_simple_graph(g):
    """Convert the graph to a simple graph with no multi-edge.

    The function generates a new *readonly* graph with no node/edge feature.

    Parameters
    ----------
    g : DGLGraph
        The input graph.

    Returns
    -------
    DGLGraph
        A simple graph.
    """
    gidx = _CAPI_DGLToSimpleGraph(g._graph)
    return DGLGraph(gidx, readonly=True)

def to_bidirected(g, readonly=True):
    """Convert the graph to a bidirected graph.

    The function generates a new graph with no node/edge feature.
    If g has an edge for i->j but no edge for j->i, then the
    returned graph will have both i->j and j->i.

    If the input graph is a multigraph (there are multiple edges from node i to node j),
    the returned graph isn't well defined.

    Parameters
    ----------
    g : DGLGraph
        The input graph.
    readonly : bool, default to be True
        Whether the returned bidirected graph is readonly or not.

    Notes
    -----
    Please make sure g is a single graph, otherwise the return value is undefined.

    Returns
    -------
    DGLGraph

    Examples
    --------
    The following two examples use PyTorch backend, one for non-multi graph
    and one for multi-graph.

    >>> g = dgl.DGLGraph()
    >>> g.add_nodes(2)
    >>> g.add_edges([0, 0], [0, 1])
    >>> bg1 = dgl.to_bidirected(g)
    >>> bg1.edges()
    (tensor([0, 1, 0]), tensor([0, 0, 1]))
    """
    if readonly:
        newgidx = _CAPI_DGLToBidirectedImmutableGraph(g._graph)
    else:
        newgidx = _CAPI_DGLToBidirectedMutableGraph(g._graph)
    return DGLGraph(newgidx)

def laplacian_lambda_max(g):
    """Return the largest eigenvalue of the normalized symmetric laplacian of g.

    The eigenvalue of the normalized symmetric of any graph is less than or equal to 2,
    ref: https://en.wikipedia.org/wiki/Laplacian_matrix#Properties

    Parameters
    ----------
    g : DGLGraph
        The input graph, it should be an undirected graph.

    Returns
    -------
    list :
        Return a list, where the i-th item indicates the largest eigenvalue
        of i-th graph in g.

    Examples
    --------

    >>> import dgl
    >>> g = dgl.DGLGraph()
    >>> g.add_nodes(5)
    >>> g.add_edges([0, 1, 2, 3, 4, 0, 1, 2, 3, 4], [1, 2, 3, 4, 0, 4, 0, 1, 2, 3])
    >>> dgl.laplacian_lambda_max(g)
    [1.809016994374948]
    """
    g_arr = unbatch(g)
    rst = []
    for g_i in g_arr:
        n = g_i.number_of_nodes()
        adj = g_i.adjacency_matrix_scipy(return_edge_ids=False).astype(float)
        norm = sparse.diags(F.asnumpy(g_i.in_degrees()).clip(1) ** -0.5, dtype=float)
        laplacian = sparse.eye(n) - norm * adj * norm
        rst.append(sparse.linalg.eigs(laplacian, 1, which='LM',
                                      return_eigenvectors=False)[0].real)
    return rst

def metapath_reachable_graph(g, metapath):
    """Return a graph where the successors of any node ``u`` are nodes reachable from ``u`` by
    the given metapath.

    If the beginning node type ``s`` and ending node type ``t`` are the same, it will return
    a homogeneous graph with node type ``s = t``.  Otherwise, a unidirectional bipartite graph
    with source node type ``s`` and destination node type ``t`` is returned.

    In both cases, two nodes ``u`` and ``v`` will be connected with an edge ``(u, v)`` if
    there exists one path matching the metapath from ``u`` to ``v``.

    The result graph keeps the node set of type ``s`` and ``t`` in the original graph even if
    they might have no neighbor.

    The features of the source/destination node type in the original graph would be copied to
    the new graph.

    Parameters
    ----------
    g : DGLHeteroGraph
        The input graph
    metapath : list[str or tuple of str]
        Metapath in the form of a list of edge types

    Returns
    -------
    DGLHeteroGraph
        A homogeneous or bipartite graph.
    """
    adj = 1
    index_dtype = g._idtype_str
    for etype in metapath:
        adj = adj * g.adj(etype=etype, scipy_fmt='csr', transpose=True)

    adj = (adj != 0).tocsr()
    srctype = g.to_canonical_etype(metapath[0])[0]
    dsttype = g.to_canonical_etype(metapath[-1])[2]
    if srctype == dsttype:
        assert adj.shape[0] == adj.shape[1]
        new_g = graph(adj, ntype=srctype, index_dtype=index_dtype)
    else:
        new_g = bipartite(adj, utype=srctype, vtype=dsttype, index_dtype=index_dtype)

    for key, value in g.nodes[srctype].data.items():
        new_g.nodes[srctype].data[key] = value
    if srctype != dsttype:
        for key, value in g.nodes[dsttype].data.items():
            new_g.nodes[dsttype].data[key] = value

    return new_g

def add_self_loop(g):
    """Return a new graph containing all the edges in the input graph plus self loops
    of every nodes.
    No duplicate self loop will be added for nodes already having self loops.
    Self-loop edges id are not preserved. All self-loop edges would be added at the end.

    Examples
    ---------

    >>> g = DGLGraph()
    >>> g.add_nodes(5)
    >>> g.add_edges([0, 1, 2], [1, 1, 2])
    >>> new_g = dgl.transform.add_self_loop(g) # Nodes 0, 3, 4 don't have self-loop
    >>> new_g.edges()
    (tensor([0, 0, 1, 2, 3, 4]), tensor([1, 0, 1, 2, 3, 4]))

    Parameters
    ------------
    g: DGLGraph

    Returns
    --------
    DGLGraph
    """
    new_g = DGLGraph()
    new_g.add_nodes(g.number_of_nodes())
    src, dst = g.all_edges(order="eid")
    src = F.zerocopy_to_numpy(src)
    dst = F.zerocopy_to_numpy(dst)
    non_self_edges_idx = src != dst
    nodes = np.arange(g.number_of_nodes())
    new_g.add_edges(src[non_self_edges_idx], dst[non_self_edges_idx])
    new_g.add_edges(nodes, nodes)
    return new_g

def remove_self_loop(g):
    """Return a new graph with all self-loop edges removed

    Examples
    ---------

    >>> g = DGLGraph()
    >>> g.add_nodes(5)
    >>> g.add_edges([0, 1, 2], [1, 1, 2])
    >>> new_g = dgl.transform.remove_self_loop(g) # Nodes 1, 2 have self-loop
    >>> new_g.edges()
    (tensor([0]), tensor([1]))

    Parameters
    ------------
    g: DGLGraph

    Returns
    --------
    DGLGraph
    """
    new_g = DGLGraph()
    new_g.add_nodes(g.number_of_nodes())
    src, dst = g.all_edges(order="eid")
    src = F.zerocopy_to_numpy(src)
    dst = F.zerocopy_to_numpy(dst)
    non_self_edges_idx = src != dst
    new_g.add_edges(src[non_self_edges_idx], dst[non_self_edges_idx])
    return new_g

def reorder_nodes(g, new_node_ids):
    """ Generate a new graph with new node Ids.

    We assign each node in the input graph with a new node Id. This results in
    a new graph.

    Parameters
    ----------
    g : DGLGraph
        The input graph
    new_node_ids : a tensor
        The new node Ids
    Returns
    -------
    DGLGraph
        The graph with new node Ids.
    """
    assert len(new_node_ids) == g.number_of_nodes(), \
            "The number of new node ids must match #nodes in the graph."
    new_node_ids = utils.toindex(new_node_ids)
    sorted_ids, idx = F.sort_1d(new_node_ids.tousertensor())
    assert F.asnumpy(sorted_ids[0]) == 0 \
            and F.asnumpy(sorted_ids[-1]) == g.number_of_nodes() - 1, \
            "The new node Ids are incorrect."
    new_gidx = _CAPI_DGLReorderGraph(g._graph, new_node_ids.todgltensor())
    new_g = DGLGraph(new_gidx)
    new_g.ndata['orig_id'] = idx
    return new_g

def partition_graph_with_halo(g, node_part, extra_cached_hops, reshuffle=False):
    '''Partition a graph.

    Based on the given node assignments for each partition, the function splits
    the input graph into subgraphs. A subgraph may contain HALO nodes which does
    not belong to the partition of a subgraph but are connected to the nodes
    in the partition within a fixed number of hops.

    If `reshuffle` is turned on, the function reshuffles node Ids and edge Ids
    of the input graph before partitioning. After reshuffling, all nodes and edges
    in a partition fall in a contiguous Id range in the input graph.
    The partitioend subgraphs have node data 'orig_id', which stores the node Ids
    in the original input graph.

    Parameters
    ------------
    g: DGLGraph
        The graph to be partitioned

    node_part: 1D tensor
        Specify which partition a node is assigned to. The length of this tensor
        needs to be the same as the number of nodes of the graph. Each element
        indicates the partition Id of a node.

    extra_cached_hops: int
        The number of hops a HALO node can be accessed.

    reshuffle : bool
        Resuffle nodes so that nodes in the same partition are in the same Id range.

    Returns
    --------
    a dict of DGLGraphs
        The key is the partition Id and the value is the DGLGraph of the partition.
    '''
    assert len(node_part) == g.number_of_nodes()
    node_part = utils.toindex(node_part)
    if reshuffle:
        node_part = node_part.tousertensor()
        sorted_part, new2old_map = F.sort_1d(node_part)
        new_node_ids = F.gather_row(F.arange(0, g.number_of_nodes()), new2old_map)
        g = reorder_nodes(g, new_node_ids)
        node_part = utils.toindex(sorted_part)
        # We reassign edges in in-CSR. In this way, after partitioning, we can ensure
        # that all edges in a partition are in the contiguous Id space.
        orig_eids = _CAPI_DGLReassignEdges(g._graph, True)
        orig_eids = utils.toindex(orig_eids)
        g.edata['orig_id'] = orig_eids.tousertensor()

    subgs = _CAPI_DGLPartitionWithHalo(g._graph, node_part.todgltensor(), extra_cached_hops)
    subg_dict = {}
    node_part = node_part.tousertensor()
    for i, subg in enumerate(subgs):
        inner_node = _get_halo_subgraph_inner_node(subg)
        subg = g._create_subgraph(subg, subg.induced_nodes, subg.induced_edges)
        inner_node = F.zerocopy_from_dlpack(inner_node.to_dlpack())
        subg.ndata['inner_node'] = inner_node
        subg.ndata['part_id'] = F.gather_row(node_part, subg.parent_nid)
        if reshuffle:
            subg.ndata['orig_id'] = F.gather_row(g.ndata['orig_id'], subg.ndata[NID])
            subg.edata['orig_id'] = F.gather_row(g.edata['orig_id'], subg.edata[EID])

        if extra_cached_hops >= 1:
            inner_edge = F.zeros((subg.number_of_edges(),), F.int64, F.cpu())
            inner_nids = F.nonzero_1d(subg.ndata['inner_node'])
            # TODO(zhengda) we need to fix utils.toindex() to avoid the dtype cast below.
            inner_nids = F.astype(inner_nids, F.int64)
            inner_eids = subg.in_edges(inner_nids, form='eid')
            inner_edge = F.scatter_row(inner_edge, inner_eids,
                                       F.ones((len(inner_eids),), F.dtype(inner_edge), F.cpu()))
        else:
            inner_edge = F.ones((subg.number_of_edges(),), F.int64, F.cpu())
        subg.edata['inner_edge'] = inner_edge
        subg_dict[i] = subg
    return subg_dict

def metis_partition_assignment(g, k):
    ''' This assigns nodes to different partitions with Metis partitioning algorithm.

    After the partition assignment, we construct partitions.

    Parameters
    ----------
    g : DGLGraph
        The graph to be partitioned
    k : int
        The number of partitions.

    Returns
    -------
    a 1-D tensor
        A vector with each element that indicates the partition Id of a vertex.
    '''
    # METIS works only on symmetric graphs.
    # The METIS runs on the symmetric graph to generate the node assignment to partitions.
    sym_g = to_bidirected(g, readonly=True)
    node_part = _CAPI_DGLMetisPartition(sym_g._graph, k)
    if len(node_part) == 0:
        return None
    else:
        node_part = utils.toindex(node_part)
        return node_part.tousertensor()

def metis_partition(g, k, extra_cached_hops=0, reshuffle=False):
    ''' This is to partition a graph with Metis partitioning.

    Metis assigns vertices to partitions. This API constructs subgraphs with the vertices assigned
    to the partitions and their incoming edges. A subgraph may contain HALO nodes which does
    not belong to the partition of a subgraph but are connected to the nodes
    in the partition within a fixed number of hops.

    If `reshuffle` is turned on, the function reshuffles node Ids and edge Ids
    of the input graph before partitioning. After reshuffling, all nodes and edges
    in a partition fall in a contiguous Id range in the input graph.
    The partitioend subgraphs have node data 'orig_id', which stores the node Ids
    in the original input graph.

    The partitioned subgraph is stored in DGLGraph. The DGLGraph has the `part_id`
    node data that indicates the partition a node belongs to. The subgraphs do not contain
    the node/edge data in the input graph.

    Parameters
    ------------
    g: DGLGraph
        The graph to be partitioned

    k: int
        The number of partitions.

    extra_cached_hops: int
        The number of hops a HALO node can be accessed.

    reshuffle : bool
        Resuffle nodes so that nodes in the same partition are in the same Id range.

    Returns
    --------
    a dict of DGLGraphs
        The key is the partition Id and the value is the DGLGraph of the partition.
    '''
    # METIS works only on symmetric graphs.
    # The METIS runs on the symmetric graph to generate the node assignment to partitions.
    sym_g = to_bidirected(g, readonly=True)
    node_part = _CAPI_DGLMetisPartition(sym_g._graph, k)
    if len(node_part) == 0:
        return None

    # Then we split the original graph into parts based on the METIS partitioning results.
    return partition_graph_with_halo(g, node_part, extra_cached_hops, reshuffle)

def compact_graphs(graphs, always_preserve=None):
    """Given a list of graphs with the same set of nodes, find and eliminate the common
    isolated nodes across all graphs.

    This function requires the graphs to have the same set of nodes (i.e. the node types
    must be the same, and the number of nodes of each node type must be the same).  The
    metagraph does not have to be the same.

    It finds all the nodes that have zero in-degree and zero out-degree in all the given
    graphs, and eliminates them from all the graphs.

    Useful for graph sampling where we have a giant graph but we only wish to perform
    message passing on a smaller graph with a (tiny) subset of nodes.

    The node and edge features are not preserved.

    Parameters
    ----------
    graphs : DGLHeteroGraph or list[DGLHeteroGraph]
        The graph, or list of graphs
    always_preserve : Tensor or dict[str, Tensor], optional
        If a dict of node types and node ID tensors is given, the nodes of given
        node types would not be removed, regardless of whether they are isolated.
        If a Tensor is given, assume that all the graphs have one (same) node type.

    Returns
    -------
    DGLHeteroGraph or list[DGLHeteroGraph]
        The compacted graph or list of compacted graphs.

        Each returned graph would have a feature ``dgl.NID`` containing the mapping
        of node IDs for each type from the compacted graph(s) to the original graph(s).
        Note that the mapping is the same for all the compacted graphs.

    Bugs
    ----
    This function currently requires that the same node type of all graphs should have
    the same node type ID, i.e. the node types are *ordered* the same.

    Examples
    --------
    The following code constructs a bipartite graph with 20 users and 10 games, but
    only user #1 and #3, as well as game #3 and #5, have connections:

    >>> g = dgl.bipartite([(1, 3), (3, 5)], 'user', 'plays', 'game', num_nodes=(20, 10))

    The following would compact the graph above to another bipartite graph with only
    two users and two games.

    >>> new_g, induced_nodes = dgl.compact_graphs(g)
    >>> induced_nodes
    {'user': tensor([1, 3]), 'game': tensor([3, 5])}

    The mapping tells us that only user #1 and #3 as well as game #3 and #5 are kept.
    Furthermore, the first user and second user in the compacted graph maps to
    user #1 and #3 in the original graph.  Games are similar.

    One can verify that the edge connections are kept the same in the compacted graph.

    >>> new_g.edges(form='all', order='eid', etype='plays')
    (tensor([0, 1]), tensor([0, 1]), tensor([0, 1]))

    When compacting multiple graphs, nodes that do not have any connections in any
    of the given graphs are removed.  So if we compact ``g`` and the following ``g2``
    graphs together:

    >>> g2 = dgl.bipartite([(1, 6), (6, 8)], 'user', 'plays', 'game', num_nodes=(20, 10))
    >>> (new_g, new_g2), induced_nodes = dgl.compact_graphs([g, g2])
    >>> induced_nodes
    {'user': tensor([1, 3, 6]), 'game': tensor([3, 5, 6, 8])}

    Then one can see that user #1 from both graphs, users #3 from the first graph, as
    well as user #6 from the second graph, are kept.  Games are similar.

    Similarly, one can also verify the connections:

    >>> new_g.edges(form='all', order='eid', etype='plays')
    (tensor([0, 1]), tensor([0, 1]), tensor([0, 1]))
    >>> new_g2.edges(form='all', order='eid', etype='plays')
    (tensor([0, 2]), tensor([2, 3]), tensor([0, 1]))
    """
    return_single = False
    if not isinstance(graphs, Iterable):
        graphs = [graphs]
        return_single = True
    if len(graphs) == 0:
        return []

    # Ensure the node types are ordered the same.
    # TODO(BarclayII): we ideally need to remove this constraint.
    ntypes = graphs[0].ntypes
    graph_dtype = graphs[0]._idtype_str
    graph_ctx = graphs[0]._graph.ctx
    for g in graphs:
        assert ntypes == g.ntypes, \
            ("All graphs should have the same node types in the same order, got %s and %s" %
             ntypes, g.ntypes)
        assert graph_dtype == g._idtype_str, "Expect graph data type to be {}, but got {}".format(
            graph_dtype, g._idtype_str)
        assert graph_ctx == g._graph.ctx, "Expect graph device to be {}, but got {}".format(
            graph_ctx, g._graph.ctx)

    # Process the dictionary or tensor of "always preserve" nodes
    if always_preserve is None:
        always_preserve = {}
    elif not isinstance(always_preserve, Mapping):
        if len(ntypes) > 1:
            raise ValueError("Node type must be given if multiple node types exist.")
        always_preserve = {ntypes[0]: always_preserve}

    always_preserve_nd = []
    for ntype in ntypes:
        nodes = always_preserve.get(ntype, None)
        if nodes is None:
            nodes = nd.empty([0], graph_dtype, graph_ctx)
        else:
            nodes = F.zerocopy_to_dgl_ndarray(nodes)
        always_preserve_nd.append(nodes)

    # Compact and construct heterographs
    new_graph_indexes, induced_nodes = _CAPI_DGLCompactGraphs(
        [g._graph for g in graphs], always_preserve_nd)
    induced_nodes = [F.zerocopy_from_dgl_ndarray(nodes.data) for nodes in induced_nodes]

    new_graphs = [
        DGLHeteroGraph(new_graph_index, graph.ntypes, graph.etypes)
        for new_graph_index, graph in zip(new_graph_indexes, graphs)]
    for g in new_graphs:
        for i, ntype in enumerate(graphs[0].ntypes):
            g.nodes[ntype].data[NID] = induced_nodes[i]
    if return_single:
        new_graphs = new_graphs[0]

    return new_graphs

def to_block(g, dst_nodes=None, include_dst_in_src=True):
    """Convert a graph into a bipartite-structured "block" for message passing.

    A block graph is uni-directional bipartite graph consisting of two sets of nodes
    SRC and DST. Each set can have many node types while all the edges are from SRC
    nodes to DST nodes.

    Specifically, for each relation graph of canonical edge type ``(utype, etype, vtype)``,
    node type ``utype`` belongs to SRC while ``vtype`` belongs to DST.
    Edges from node type ``utype`` to node type ``vtype`` are preserved. If
    ``utype == vtype``, the result graph will have two node types of the same name ``utype``,
    but one belongs to SRC while the other belongs to DST. This is because although
    they have the same name, their node ids are relabeled differently (see below). In
    both cases, the canonical edge type in the new graph is still
    ``(utype, etype, vtype)``, so there is no difference when referring to it.

    Moreover, the function also relabels node ids in each type to make the graph more compact.
    Specifically, the nodes of type ``vtype`` would contain the nodes that have at least one
    inbound edge of any type, while ``utype`` would contain all the DST nodes of type ``vtype``,
    as well as the nodes that have at least one outbound edge to any DST node.

    Since DST nodes are included in SRC nodes, a common requirement is to fetch
    the DST node features from the SRC nodes features. To avoid expensive sparse lookup,
    the function assures that the DST nodes in both SRC and DST sets have the same ids.
    As a result, given the node feature tensor ``X`` of type ``utype``,
    the following code finds the corresponding DST node features of type ``vtype``:

    .. code::

        X[:block.number_of_nodes('DST/vtype')]

    If the ``dst_nodes`` argument is given, the DST nodes would contain the given nodes.
    Otherwise, the DST nodes would be determined by DGL via the rules above.

    Parameters
    ----------
    graph : DGLHeteroGraph
        The graph.
    dst_nodes : Tensor or dict[str, Tensor], optional
        Optional DST nodes. If a tensor is given, the graph must have only one node type.
    include_dst_in_src : bool, default True
        If False, do not include DST nodes in SRC nodes.

    Returns
    -------
    DGLHeteroGraph
        The new graph describing the block.

        The node IDs induced for each type in both sides would be stored in feature
        ``dgl.NID``.

        The edge IDs induced for each type would be stored in feature ``dgl.EID``.

    Notes
    -----
    This function is primarily for creating the structures for efficient
    computation of message passing.  See [TODO] for a detailed example.

    Examples
    --------
    Converting a homogeneous graph to a block as described above:
    >>> g = dgl.graph([(0, 1), (1, 2), (2, 3)])
    >>> block = dgl.to_block(g, torch.LongTensor([3, 2]))

    The right hand side nodes would be exactly the same as the ones given: [3, 2].
    >>> induced_dst = block.dstdata[dgl.NID]
    >>> induced_dst
    tensor([3, 2])

    The first few nodes of the left hand side nodes would also be exactly the same as
    the ones given.  The rest of the nodes are the ones necessary for message passing
    into nodes 3, 2.  This means that the node 1 would be included.
    >>> induced_src = block.srcdata[dgl.NID]
    >>> induced_src
    tensor([3, 2, 1])

    We can notice that the first two nodes are identical to the given nodes as well as
    the right hand side nodes.

    The induced edges can also be obtained by the following:
    >>> block.edata[dgl.EID]
    tensor([2, 1])

    This indicates that edge (2, 3) and (1, 2) are included in the result graph.  We can
    verify that the first edge in the block indeed maps to the edge (2, 3), and the
    second edge in the block indeed maps to the edge (1, 2):
    >>> src, dst = block.edges(order='eid')
    >>> induced_src[src], induced_dst[dst]
    (tensor([2, 1]), tensor([3, 2]))

    Converting a heterogeneous graph to a block is similar, except that when specifying
    the right hand side nodes, you have to give a dict:
    >>> g = dgl.bipartite([(0, 1), (1, 2), (2, 3)], utype='A', vtype='B')

    If you don't specify any node of type A on the right hand side, the node type ``A``
    in the block would have zero nodes on the DST side.
    >>> block = dgl.to_block(g, {'B': torch.LongTensor([3, 2])})
    >>> block.number_of_dst_nodes('A')
    0
    >>> block.number_of_dst_nodes('B')
    2
    >>> block.dstnodes['B'].data[dgl.NID]
    tensor([3, 2])

    The left hand side would contain all the nodes on the right hand side:
    >>> block.srcnodes['B'].data[dgl.NID]
    tensor([3, 2])

    As well as all the nodes that have connections to the nodes on the right hand side:
    >>> block.srcnodes['A'].data[dgl.NID]
    tensor([2, 1])
    """
    if dst_nodes is None:
        # Find all nodes that appeared as destinations
        dst_nodes = defaultdict(list)
        for etype in g.canonical_etypes:
            _, dst = g.edges(etype=etype)
            dst_nodes[etype[2]].append(dst)
        dst_nodes = {ntype: F.unique(F.cat(values, 0)) for ntype, values in dst_nodes.items()}
    elif not isinstance(dst_nodes, Mapping):
        # dst_nodes is a Tensor, check if the g has only one type.
        if len(g.ntypes) > 1:
            raise ValueError(
                'Graph has more than one node type; please specify a dict for dst_nodes.')
        dst_nodes = {g.ntypes[0]: dst_nodes}

    # dst_nodes is now a dict
    dst_nodes_nd = []
    for ntype in g.ntypes:
        nodes = dst_nodes.get(ntype, None)
        if nodes is not None:
            dst_nodes_nd.append(F.zerocopy_to_dgl_ndarray(nodes))
        else:
            dst_nodes_nd.append(nd.NULL[g._idtype_str])

    new_graph_index, src_nodes_nd, induced_edges_nd = _CAPI_DGLToBlock(
        g._graph, dst_nodes_nd, include_dst_in_src)

    # The new graph duplicates the original node types to SRC and DST sets.
    new_ntypes = ([ntype for ntype in g.ntypes], [ntype for ntype in g.ntypes])
    new_graph = DGLHeteroGraph(new_graph_index, new_ntypes, g.etypes)
    assert new_graph.is_unibipartite  # sanity check

    for i, ntype in enumerate(g.ntypes):
        new_graph.srcnodes[ntype].data[NID] = F.zerocopy_from_dgl_ndarray(src_nodes_nd[i].data)
        if ntype in dst_nodes:
            new_graph.dstnodes[ntype].data[NID] = dst_nodes[ntype]
        else:
            # For empty dst node sets, still create empty mapping arrays.
            new_graph.dstnodes[ntype].data[NID] = F.tensor([], dtype=g.idtype)

    for i, canonical_etype in enumerate(g.canonical_etypes):
        induced_edges = F.zerocopy_from_dgl_ndarray(induced_edges_nd[i].data)
        utype, etype, vtype = canonical_etype
        new_canonical_etype = (utype, etype, vtype)
        new_graph.edges[new_canonical_etype].data[EID] = induced_edges

    return new_graph

def remove_edges(g, edge_ids):
    """Return a new graph with given edge IDs removed.

    The nodes are preserved.

    Parameters
    ----------
    graph : DGLHeteroGraph
        The graph
    edge_ids : Tensor or dict[etypes, Tensor]
        The edge IDs for each edge type.

    Returns
    -------
    DGLHeteroGraph
        The new graph.
        The edge ID mapping from the new graph to the original graph is stored as
        ``dgl.EID`` on edge features.
    """
    if not isinstance(edge_ids, Mapping):
        if len(g.etypes) != 1:
            raise ValueError(
                "Graph has more than one edge type; specify a dict for edge_id instead.")
        edge_ids = {g.canonical_etypes[0]: edge_ids}

    edge_ids_nd = [nd.NULL[g._idtype_str]] * len(g.etypes)
    for key, value in edge_ids.items():
        if value.dtype != g.idtype:
            # if didn't check, this function still works, but returns wrong result
            raise utils.InconsistentDtypeException("Expect edge id tensors({}) to have \
         the same index type as graph({})".format(value.dtype, g.idtype))
        edge_ids_nd[g.get_etype_id(key)] = F.zerocopy_to_dgl_ndarray(value)
    new_graph_index, induced_eids_nd = _CAPI_DGLRemoveEdges(g._graph, edge_ids_nd)

    new_graph = DGLHeteroGraph(new_graph_index, g.ntypes, g.etypes)
    for i, canonical_etype in enumerate(g.canonical_etypes):
        data = induced_eids_nd[i].data
        if len(data) == 0:
            # Empty means that either
            # (1) no edges are removed and edges are not shuffled.
            # (2) all edges are removed.
            # The following statement deals with both cases.
            new_graph.edges[canonical_etype].data[EID] = F.arange(
                0, new_graph.number_of_edges(canonical_etype))
        else:
            new_graph.edges[canonical_etype].data[EID] = F.zerocopy_from_dgl_ndarray(data)

    return new_graph

def in_subgraph(g, nodes):
    """Extract the subgraph containing only the in edges of the given nodes.

    The subgraph keeps the same type schema and the cardinality of the original one.
    Node/edge features are not preserved. The original IDs
    the extracted edges are stored as the `dgl.EID` feature in the returned graph.

    Parameters
    ----------
    g : DGLHeteroGraph
        Full graph structure.
    nodes : tensor or dict
        Node ids to sample neighbors from. The allowed types
        are dictionary of node types to node id tensors, or simply node id tensor if
        the given graph g has only one type of nodes.

    Returns
    -------
    DGLHeteroGraph
        The subgraph.
    """
    if not isinstance(nodes, dict):
        if len(g.ntypes) > 1:
            raise DGLError("Must specify node type when the graph is not homogeneous.")
        nodes = {g.ntypes[0] : nodes}
    nodes_all_types = []
    for ntype in g.ntypes:
        if ntype in nodes:
            nodes_all_types.append(utils.toindex(nodes[ntype], g._idtype_str).todgltensor())
        else:
            nodes_all_types.append(nd.NULL[g._idtype_str])

    subgidx = _CAPI_DGLInSubgraph(g._graph, nodes_all_types)
    induced_edges = subgidx.induced_edges
    ret = DGLHeteroGraph(subgidx.graph, g.ntypes, g.etypes)
    for i, etype in enumerate(ret.canonical_etypes):
        ret.edges[etype].data[EID] = induced_edges[i].tousertensor()
    return ret

def out_subgraph(g, nodes):
    """Extract the subgraph containing only the out edges of the given nodes.

    The subgraph keeps the same type schema and the cardinality of the original one.
    Node/edge features are not preserved. The original IDs
    the extracted edges are stored as the `dgl.EID` feature in the returned graph.

    Parameters
    ----------
    g : DGLHeteroGraph
        Full graph structure.
    nodes : tensor or dict
        Node ids to sample neighbors from. The allowed types
        are dictionary of node types to node id tensors, or simply node id tensor if
        the given graph g has only one type of nodes.

    Returns
    -------
    DGLHeteroGraph
        The subgraph.
    """
    if not isinstance(nodes, dict):
        if len(g.ntypes) > 1:
            raise DGLError("Must specify node type when the graph is not homogeneous.")
        nodes = {g.ntypes[0] : nodes}
    nodes_all_types = []
    for ntype in g.ntypes:
        if ntype in nodes:
            nodes_all_types.append(utils.toindex(nodes[ntype], g._idtype_str).todgltensor())
        else:
            nodes_all_types.append(nd.NULL[g._idtype_str])

    subgidx = _CAPI_DGLOutSubgraph(g._graph, nodes_all_types)
    induced_edges = subgidx.induced_edges
    ret = DGLHeteroGraph(subgidx.graph, g.ntypes, g.etypes)
    for i, etype in enumerate(ret.canonical_etypes):
        ret.edges[etype].data[EID] = induced_edges[i].tousertensor()
    return ret

def to_simple(g, return_counts='count', writeback_mapping=None):
    """Convert a heterogeneous multigraph to a heterogeneous simple graph, coalescing
    duplicate edges into one.

    This function does not preserve node and edge features.

    Parameters
    ----------
    g : DGLHeteroGraph
        The heterogeneous graph
    return_counts : str, optional
        If given, the returned graph would have a column with the same name that stores
        the number of duplicated edges from the original graph.
    writeback_mapping : str, optional
        If given, the mapping from the edge IDs of original graph to those of the returned
        graph would be written into edge feature with this name in the original graph for
        each edge type.

    Returns
    -------
    DGLHeteroGraph
        The new heterogeneous simple graph.

    Examples
    --------
    Consider the following graph
    >>> g = dgl.graph([(0, 1), (1, 3), (2, 2), (1, 3), (1, 4), (1, 4)])
    >>> sg = dgl.to_simple(g, return_counts='weights', writeback_mapping='new_eid')

    The returned graph would have duplicate edges connecting (1, 3) and (1, 4) removed:
    >>> sg.all_edges(form='uv', order='eid')
    (tensor([0, 1, 1, 2]), tensor([1, 3, 4, 2]))

    If ``return_counts`` is set, the returned graph will also return how many edges
    in the original graph are connecting the endpoints of the edges in the new graph:
    >>> sg.edata['weights']
    tensor([1, 2, 2, 1])

    This essentially reads that one edge is connecting (0, 1) in ``g``, whereas 2 edges
    are connecting (1, 3) in ``g``, etc.

    One can also retrieve the mapping from the edges in the original graph to edges in
    the new graph by setting ``writeback_mapping`` and running
    >>> g.edata['new_eid']
    tensor([0, 1, 3, 1, 2, 2])

    This tells us that the first edge in ``g`` is mapped to the first edge in ``sg``, and
    the second and the fourth edge are mapped to the second edge in ``sg``, etc.
    """
    simple_graph_index, counts, edge_maps = _CAPI_DGLToSimpleHetero(g._graph)
    simple_graph = DGLHeteroGraph(simple_graph_index, g.ntypes, g.etypes)
    counts = [F.zerocopy_from_dgl_ndarray(count.data) for count in counts]
    edge_maps = [F.zerocopy_from_dgl_ndarray(edge_map.data) for edge_map in edge_maps]

    if return_counts is not None:
        for count, canonical_etype in zip(counts, g.canonical_etypes):
            simple_graph.edges[canonical_etype].data[return_counts] = count

    if writeback_mapping is not None:
        for edge_map, canonical_etype in zip(edge_maps, g.canonical_etypes):
            g.edges[canonical_etype].data[writeback_mapping] = edge_map

    return simple_graph

def as_heterograph(g, ntype='_U', etype='_E'):
    """Convert a DGLGraph to a DGLHeteroGraph with one node and edge type.

    Node and edge features are preserved. Returns 64 bits graph

    Parameters
    ----------
    g : DGLGraph
        The graph
    ntype : str, optional
        The node type name
    etype : str, optional
        The edge type name

    Returns
    -------
    DGLHeteroGraph
        The heterograph.
    """
    hgi = _CAPI_DGLAsHeteroGraph(g._graph)
    hg = DGLHeteroGraph(hgi, [ntype], [etype])
    hg.ndata.update(g.ndata)
    hg.edata.update(g.edata)
    return hg

def as_immutable_graph(hg):
    """Convert a DGLHeteroGraph with one node and edge type into a DGLGraph.

    Node and edge features are preserved.

    Parameters
    ----------
    g : DGLHeteroGraph
        The heterograph

    Returns
    -------
    DGLGraph
        The graph.
    """
    gidx = _CAPI_DGLAsImmutableGraph(hg._graph)
    g = DGLGraph(gidx)
    g.ndata.update(hg.ndata)
    g.edata.update(hg.edata)
    return g

def sort_csr_(g, tag=None, split="_SPLIT"):
    """Sort the (out)CSR matrix of the graph inplace

    After sorting, edges whose destination shares the same tag will be arranged in
    a consecutive range. The COO matrix, in CSR and edge ids remain unchanged.
    Currently, it only support homograph and bipartite graph.

    Parameters
    ----------
    g : DGLHeteroGraph
        The heterograph
    tag : str
        The name of the node feature used as tag
        When tag is None, sort by the node id of destination nodes.
    split: str
        The name of the node feature to store split positions of different tags in the
        adjancency list.

        For example, if the adjancy list looks like this after sorting:
            dst: 0, 1, 2, 3, 4, 5, 6, 7, 8
            tag: 0, 0, 1, 1, 2, 2, 2, 3, 3
                ^    ^     ^        ^     ^
        Then ndata[split][node_id] will be:
            [0, 2, 4, 7, 9] (Marked with ^)
    """
    if len(g.etypes) > 1:
        raise DGLError("Only support homograph and bipartite graph")
    srctype = g.srctypes[0]
    dsttype = g.dsttypes[0]
    if tag is None:
        tag_arr = nd.NULL["int32"]
        num_tags = 0
    else:
        tag_data = g.nodes[dsttype].data[tag]
        num_tags = int(F.asnumpy(F.max(tag_data, 0))) + 1
        tag_arr = F.zerocopy_to_dgl_ndarray(tag_data)
    ret = _CAPI_DGLHeteroSortCSR_(g._graph, 0, tag_arr, num_tags)
    if tag is not None:
        ret = F.zerocopy_from_dgl_ndarray(ret)
        g.nodes[srctype].data[split] = F.reshape(ret, [-1, num_tags + 1])

def sort_csc_(g, tag=None, split="_SPLIT"):
    """Sort the CSC(in CSR) matrix of the graph inplace

    After sorting, edges whose source shares the same tag will be arranged in
    a consecutive range. The COO matrix, out CSR and edge ids remain unchanged.
    Currently, it only support homograph and bipartite graph.

    Parameters
    ----------
    g : DGLHeteroGraph
        The heterograph
    tag : str
        The name of the node feature used as tag
        When tag is None, sort by the node id of destination nodes.
    split : str
        The name of the node feature to store split positions of different tags in the
        adjancency list.

    """
    if len(g.etypes) > 1:
        raise DGLError("Only support homograph and bipartite graph")
    srctype = g.srctypes[0]
    dsttype = g.dsttypes[0]
    if tag is None:
        tag_arr = nd.NULL["int32"]
        num_tags = 0
    else:
        tag_data = g.nodes[srctype].data[tag]
        num_tags = int(F.asnumpy(F.max(tag_data, 0))) + 1
        tag_arr = F.zerocopy_to_dgl_ndarray(tag_data)
    ret = _CAPI_DGLHeteroSortCSC_(g._graph, 0, tag_arr, num_tags)
    if tag is not None:
        ret = F.zerocopy_from_dgl_ndarray(ret)
        g.nodes[dsttype].data[split] = F.reshape(ret, [-1, num_tags + 1])

def sort_csr(g, tag=None, split="_SPLIT"):
    """A copy of the given graph whose (out)CSR matrix is sorted.

    The outplace version of sort_csr_
    Node frames and edges frames are shallow copy of the original graph.

    Parameters
    ----------
    g : DGLHeteroGraph
        The heterograph
    tag : str
        The name of the node feature used as tag
        When tag is None, sort by the node id of destination nodes.
    split: str
        The name of the node feature to store split positions of different tags in the
        adjancency list.

    Returns
    -------
    DGLHeteroGraph
        A copy of the given graph whose (out)CSR matrix is sorted.

    """
    if len(g.etypes) > 1:
        raise DGLError("Only support homograph and bipartite graph")
    srctype = g.srctypes[0]
    dsttype = g.dsttypes[0]
    if tag is None:
        tag_arr = nd.NULL["int32"]
        num_tags = 0
    else:
        tag_data = g.nodes[dsttype].data[tag]
        num_tags = int(F.asnumpy(F.max(tag_data, 0))) + 1
        tag_arr = F.zerocopy_to_dgl_ndarray(tag_data)
    ret = _CAPI_DGLHeteroSortCSR(g._graph, 0, tag_arr, num_tags)
    gidx, split_arr = [item.data for item in ret]
    node_frames = [FrameRef(Frame(num_rows=self._graph.number_of_nodes(i)))
                   if frame is None else frame.clone()
                   for i, frame in enumerate(g._node_frames)]
    edge_frames = [FrameRef(Frame(num_rows=self._graph.number_of_edges(i)))
                   if frame is None else frame.clone()
                   for i, frame in enumerate(g._edge_frames)]
    new_g = DGLHeteroGraph(gidx, g.ntypes, g.etypes, node_frames, edge_frames)
    if tag is not None:
        split_arr = F.reshape(F.zerocopy_from_dgl_ndarray(split_arr), [-1, num_tags + 1])
        new_g.nodes[srctype].data[split] = split_arr
    return new_g

def sort_csc(g, tag=None, split="_SPLIT"):
    """A copy of the given graph whose CSC(in CSR) matrix is sorted.

    The outplace version of sort_csc_
    Node frames and edges frames are shallow copy of the original graph.

    Parameters
    ----------
    g : DGLHeteroGraph
        The heterograph
    tag : str
        The name of the node feature used as tag
        When tag is None, sort by the node id of destination nodes.
    split: str
        The name of the node feature to store split positions of different tags in the
        adjancency list.

    Returns
    -------
    DGLHeteroGraph
        A copy of the given graph whose CSC(in CSR) matrix is sorted.

    """
    if len(g.etypes) > 1:
        raise DGLError("Only support homograph and bipartite graph")
    srctype = g.srctypes[0]
    dsttype = g.dsttypes[0]
    if tag is None:
        tag_arr = nd.NULL["int32"]
        num_tags = 0
    else:
        tag_data = g.nodes[srctype].data[tag]
        num_tags = int(F.asnumpy(F.max(tag_data, 0))) + 1
        tag_arr = F.zerocopy_to_dgl_ndarray(tag_data)
    ret = _CAPI_DGLHeteroSortCSC(g._graph, 0, tag_arr, num_tags)
    gidx, split_arr = [item.data for item in ret]
    node_frames = [FrameRef(Frame(num_rows=self._graph.number_of_nodes(i)))
                   if frame is None else frame.clone()
                   for i, frame in enumerate(g._node_frames)]
    edge_frames = [FrameRef(Frame(num_rows=self._graph.number_of_edges(i)))
                   if frame is None else frame.clone()
                   for i, frame in enumerate(g._edge_frames)]
    new_g = DGLHeteroGraph(gidx, g.ntypes, g.etypes, node_frames, edge_frames)
    if tag is not None:
        split_arr = F.reshape(F.zerocopy_from_dgl_ndarray(split_arr), [-1, num_tags + 1])
        new_g.nodes[dsttype].data[split] = split_arr
    return new_g

_init_api("dgl.transform")