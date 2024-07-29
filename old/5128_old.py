
        # custom output channels of the classifier
        self.custom_cls_channels = True
        # custom accuracy of the classsifier
        self.custom_acc = True

    def get_cls_channels(self, num_classes):
        # cls_channels = num_classes + 2
        assert num_classes == self.num_classes
        return num_classes + 2

    def _split_cls_score(self, cls_score):
        # split cls_score to cls_score_classes and cls_score_objectness
        assert cls_score.size(-1) == self.num_classes + 2
        cls_score_classes = cls_score[..., :-2]
        cls_score_objectness = cls_score[..., -2:]
        return cls_score_classes, cls_score_objectness

    def get_activation(self, cls_score):
        # calculate the scores for detection results
        # scores = score_classes * score_objectness
        cls_score_classes, cls_score_objectness = self._split_cls_score(
            cls_score)
        score_classes = F.softmax(cls_score_classes, dim=-1)
        score_objectness = F.softmax(cls_score_objectness, dim=-1)
        score_pos = score_objectness[..., [0]]
        score_neg = score_objectness[..., [1]]
        score_classes = score_classes * score_pos
        scores = torch.cat([score_classes, score_neg], dim=-1)
        return scores

    def get_accuracy(self, cls_score, labels):
        pos_inds = labels < self.num_classes
        obj_labels = (labels == self.num_classes).long()
        cls_score_classes, cls_score_objectness = self._split_cls_score(
            cls_score)
        acc_objectness = accuracy(cls_score_objectness, obj_labels)
        acc_classes = accuracy(cls_score_classes[pos_inds], labels[pos_inds])
        acc = dict()
        acc['acc_objectness'] = acc_objectness
        acc['acc_classes'] = acc_classes
        return acc

    def forward(self,
                cls_score,
                labels,
                label_weights,
                avg_factor=None,
                reduction_override=None):
        """Forward function.

        Args:
            cls_score (torch.Tensor): The prediction with shape (N, C + 2).
            labels (torch.Tensor): The learning label of the prediction.
            label_weights (torch.Tensor, optional): Sample-wise loss weight.
            avg_factor (int, optional): Average factor that is used to average
                 the loss. Defaults to None.
            reduction (str, optional): The method used to reduce the loss.
                 Options are "none", "mean" and "sum".
        Returns:
            torch.Tensor: The calculated loss
        """
        assert reduction_override in (None, 'none', 'mean', 'sum')
        reduction = (
            reduction_override if reduction_override else self.reduction)
        assert cls_score.size(-1) == self.num_classes + 2
        pos_inds = labels < self.num_classes
        # 0 for pos, 1 for neg
        obj_labels = (labels == self.num_classes).long()

        # accumulate the samples for each category
        unique_labels = labels.unique()
        for u_l in unique_labels:
            inds_ = labels == u_l.item()
            self.cum_samples[u_l] += inds_.sum()

        if label_weights is not None:
            label_weights = label_weights.float()

        cls_score_classes, cls_score_objectness = self._split_cls_score(
            cls_score)
        # calculate loss_cls_classes (only need pos samples)
        loss_cls_classes = self.loss_weight * self.cls_criterion(
            cls_score_classes[pos_inds], labels[pos_inds],
            label_weights[pos_inds], self.cum_samples[:self.num_classes],
            self.num_classes, self.p, self.q, self.eps, reduction, avg_factor)
        # calculate loss_cls_objectness
        loss_cls_objectness = self.loss_weight * cross_entropy(
            cls_score_objectness, obj_labels, label_weights, reduction,
            avg_factor)

        if self.return_dict:
            loss_cls = dict()
            loss_cls['loss_cls_objectness'] = loss_cls_objectness
            loss_cls['loss_cls_classes'] = loss_cls_classes
        else:
            loss_cls = loss_cls_classes + loss_cls_objectness
        return loss_cls
