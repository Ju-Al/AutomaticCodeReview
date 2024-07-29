        self.num_branch = num_branch
        self.test_branch_idx = test_branch_idx
        super(TridentFasterRCNN, self).__init__(
            backbone=backbone,
            neck=neck,
            rpn_head=rpn_head,
            roi_head=roi_head,
            train_cfg=train_cfg,
            test_cfg=test_cfg,
            pretrained=pretrained)

    def simple_test(self, img, img_metas, proposals=None, rescale=False):
        """Test without augmentation."""
        assert self.with_bbox, 'Bbox head must be implemented.'

        x = self.extract_feat(img)

        if proposals is None:
            num_branch = (self.num_branch if self.test_branch_idx == -1 else 1)
            trident_img_metas = img_metas * num_branch
            proposal_list = self.rpn_head.simple_test_rpn(x, trident_img_metas)
        else:
            proposal_list = proposals

        return self.roi_head.simple_test(
            x, proposal_list, trident_img_metas, rescale=rescale)

    def forward_train(self, img, img_metas, gt_bboxes, gt_labels, **kwargs):
        """make copies of img and gts to fit multi-branch."""
        trident_gt_bboxes = tuple(gt_bboxes * self.num_branch)
        trident_gt_labels = tuple(gt_labels * self.num_branch)
        trident_img_metas = tuple(img_metas * self.num_branch)

        return super(TridentFasterRCNN,
                     self).forward_train(img, trident_img_metas,
                                         trident_gt_bboxes, trident_gt_labels)
