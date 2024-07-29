

@not_implemented_for("directed", "multigraph")
def d_fragmented(G: nx.Graph, d: int) -> Generator:
    """
    Given a simple undirected graph G, compute a
    d-club cluster graph.

    Parameters
    ----------
    G: A networkX graph instance.
        Any simple undirected graph.
    d: A non-negative int.
        The club-size of a cluster graph.

    Raises
    ------
    NetworkXNotImplemented
        The case where G is not a Forest and d > 1 is not
        supported.


    Yields
    -------
    edge: Tuple
        Tuples that represent undirected edges
        between nodes of G.


    Notes
    ______

    Runs in O(|V(G)|) if
    G is a Forest (a collection of
    pair-wise disconnected trees).
    This algorithm is provided in
    [1]_.

    If G is not a Forest and d = 1,
    then it relies on an efficient
    implementation of maximum matching
    algorithms in arbitrary graphs (like
    Edmonds' blossom algorithm, or Micali-Vazirani Matching).

    In this case, the implementation of Edmonds' blossom
    algorithm (runs in O(|V|^2 * |E|)) is used
    because it is available in networkX.

    In contrast to that, Micali-Vazirani Matching
    runs in O(sqrt(|V|) * |E|) which is significantly
    faster but the algorithm itself is extremely complicated
    to implement.


    If more results are known, it is encouraged to add to this file
    the support for more cases.

    References
    ----------
    .. [1] Oellerman, O., Patel, A.,
        The s-club deletion problem on various graph families,
        University of Winnipeg, Canada, 2020.

    """
    if d < 0:
        raise ValueError("Note d cannot be negative.")
    elif d == 0:
        yield from G.edges
    elif _is_forest(G):
        for comp in connected.connected_components(G):
            h = G.subgraph(comp)
            if len(h) > 1:
                yield from iter(_d_fragment_tree(h, d))
    elif d == 1:
        # Note M is a maximum matching in G if and only if
        # G[G-M] is a 1-fragmentation of G.
        mwm = matching.max_weight_matching(G)
        yield from iter(set(G.edges()) - set(mwm))
    else:
        # TODO: Implement d-fragmentation for arbitrary graphs

        no_good_algorithm_known = """
            If d > 1, then no fast algorithms of d-fragmentation are known for graphs that are not forests.
            In fact, the problem of extracting a maximum d-club cluster graph from an arbitrary graph
            for d >= 2 is known to be NP-Complete.

            The current implementation only supports forests for d >= 1 or general graphs for d = 1.
            If you know one, please consider adding it to NetworkX by following:

            https://github.com/networkx/networkx/blob/master/CONTRIBUTING.rst
        """
        raise nx.exception.NetworkXNotImplemented(no_good_algorithm_known)


def _is_leaf(G: nx.Graph, node):
    """
    Return True if node is a leaf,
    False otherwise.
    """

    return len(list(G.neighbors(node))) == 1


def _get_node_with_highest_degree(G: nx.Graph):
    """
    Get a random node from the vertex
    set of the graph.
    """
    if len(list(G.nodes)) == 0:
        raise ValueError("Empty graph has no nodes.")
    heaviest_node = max(list(G.nodes), key=lambda x: G.degree[x])
    return heaviest_node


def _d_fragment_tree_helper(G: nx.Graph, parent, node, sub_ecc: Dict, d: int, edge_set):
    """
    A helper for recursive computation
    of a d-club cluster graph from G.

    Parameters
    ----------
    G: networkX.graph()
        A Tree.
    parent: node
        Some vertex of G. It could be a sentinel as well if node is the root.
    node: node
        Some vertex of G. This cannot be sentinel.
    sub_ecc: dict
        A dictionary of computed eccentricities
        for subtrees rooted at some vertex.
    d: int
        A positive int.
    edge_set: Set
        A set that collects the edges that
        form an edge cut.

    Returns
    -------
    sub_ecc, edge_set: Tuple
        A tuple (sub_ecc, edge_set) where `sub_ecc`
        is a populated dictionary of computed
        eccentricities for subtrees rooted at some vertex
        and `edge_set` is a minimum edge cut.
    """
    if _is_leaf(G, node):
        sub_ecc[node] = 0
    else:
        for n in G.neighbors(node):
            if n != parent:
                _d_fragment_tree_helper(G, node, n, sub_ecc, d, edge_set)

        ecc_values = []

        for c in G.neighbors(node):
            if c in sub_ecc and c != parent:
                ecc_values.append((c, sub_ecc[c]))

        values = sorted(ecc_values, key=itemgetter(1), reverse=True)

        if len(values) == 1:
            child, child_ecc = values[0]

            if child_ecc >= d:
                edge_set |= {(child, node)}
                sub_ecc[node] = 0
            else:
                sub_ecc[node] = sub_ecc[child] + 1
        else:
            no_more = False

            for _counter in range(len(values) - 1):
                _first = values[_counter]
                _second = values[_counter + 1]

                child_first, ecc_first = _first
                child_second, ecc_second = _second

                if ecc_first + ecc_second + 2 >= d + 1:
                    edge_set |= {(child_first, node)}
                else:
                    sub_ecc[node] = sub_ecc[child_first] + 1
                    no_more = True
                    break

            if not no_more:
                child_last, ecc_last = values[-1]
                if ecc_last >= d:
                    edge_set |= {(child_last, node)}
                    sub_ecc[node] = 0
                else:
                    sub_ecc[node] = sub_ecc[child_last] + 1
    return sub_ecc, edge_set


def _d_fragment_tree(G: nx.Graph, d: int):
    """
    Compute the minimum edge cut
    that transforms G into a
    d-club cluster graph.

    Parameters
    ----------
    G: A networkX graph.
        A tree.
    d: A non-negative int
        The club-size for a cluster graph of G.

    Returns
    -------

    """
    node = _get_node_with_highest_degree(G)
    sub_ecc = dict()
    parent = -1
    edge_set = set()
    ecc, edges = _d_fragment_tree_helper(G, parent, node, sub_ecc, d, edge_set)
    return edges


# Leaving the following function commented
# so that if more d-fragmentation algorithms are added
# in future, then they can be tested locally.

# def tester():
#     adjacency = {
#         0: {1, 6},
#         1: {0, 3},
#         2: {3},
#         3: {1, 2},
#         4: {7},
#         5: {7},
#         6: {0, 7},
#         7: {4, 5, 6}
#     }
#
#     def get_edges_from_adjacency(adj):
#         return [(u, v) for u, nbrs in adj.items() for v in nbrs]
#
#     g = nx.Graph()
#     g.update(edges=get_edges_from_adjacency(adjacency), nodes=adjacency)
#
#     # print(_is_forest(G))
#     for edg in d_fragmented(g, 2):
#         print(edg)
#         pass
