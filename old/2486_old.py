            # Select Top k neighbors
            subgraph = select_topk(graph,self.k,'y')
            # Sigmoid as information threshold
            subgraph.ndata['y'] = torch.sigmoid(subgraph.ndata['y'])
            # Using vector matrix elementwise mul for acceleration
            feat = subgraph.ndata['y'].view(-1,1)*feat
            feat = self.feat_drop(feat)
            h = self.fc(feat).view(-1, self.num_heads, self.out_feats)
            el = (h * self.attn_l).sum(dim=-1).unsqueeze(-1)
            er = (h * self.attn_r).sum(dim=-1).unsqueeze(-1)
            # Assign the value on the subgraph
            subgraph.srcdata.update({'ft': h, 'el': el})
            subgraph.dstdata.update({'er': er})
            # compute edge attention, el and er are a_l Wh_i and a_r Wh_j respectively.
            subgraph.apply_edges(fn.u_add_v('el', 'er', 'e'))
            e = self.leaky_relu(subgraph.edata.pop('e'))
            # compute softmax
            subgraph.edata['a'] = self.attn_drop(edge_softmax(subgraph, e))
            # message passing
            subgraph.update_all(fn.u_mul_e('ft', 'a', 'm'),
                             fn.sum('m', 'ft'))
            rst = subgraph.dstdata['ft']
            
            # activation
            if self.activation:
                rst = self.activation(rst)
            # Residual
            if self.residual:
                rst = rst + self.residual_module(feat).view(feat.shape[0],-1,self.out_feats)

            if get_attention:
                return rst, subgraph.edata['a']
            else:
                return rst

class HardGAT(nn.Module):
    def __init__(self,
                 g,
                 num_layers,
                 in_dim,
                 num_hidden,
                 num_classes,
                 heads,
                 activation,
                 feat_drop,
                 attn_drop,
                 negative_slope,
                 residual,
                 k):
        super(HardGAT, self).__init__()
        self.g = g
        self.num_layers = num_layers
        self.gat_layers = nn.ModuleList()
        self.activation = activation
        gat_layer = partial(HardGAO,k=k)
        muls = heads
        # input projection (no residual)
        self.gat_layers.append(gat_layer(
            in_dim, num_hidden, heads[0],
            feat_drop, attn_drop, negative_slope, False, self.activation))
        # hidden layers
        for l in range(1, num_layers):
            # due to multi-head, the in_dim = num_hidden * num_heads
            self.gat_layers.append(gat_layer(
                num_hidden*muls[l-1] , num_hidden, heads[l],
                feat_drop, attn_drop, negative_slope, residual, self.activation))
        # output projection
        self.gat_layers.append(gat_layer(
            num_hidden*muls[-2] , num_classes, heads[-1],
            feat_drop, attn_drop, negative_slope, False, None))

    def forward(self, inputs):
        h = inputs
        for l in range(self.num_layers):
            h = self.gat_layers[l](self.g, h).flatten(1)
        logits = self.gat_layers[-1](self.g, h).mean(1)
        return logits
