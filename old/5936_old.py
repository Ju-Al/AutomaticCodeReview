    for anchor in tmp_res:
        cls_set[model.CLASSES.index(anchor['class_name'])]\
            .append([*anchor['bbox'], anchor['score']])
    result = []
    for cls in cls_set:
        if len(cls) == 0:
            result.append(np.array(cls, dtype=np.float32).reshape((0, 5)))
        else:
            result.append(np.array(cls, dtype=np.float32))
    return result


def main(args):
    # build the model from a config file and a checkpoint file
    model = init_detector(args.config, args.checkpoint, device=args.device)
    # test a single image
    result = inference_detector(model, args.img)
    # show the results
    show_result_pyplot(
        model,
        args.img,
        result,
        score_thr=args.score_thr,
        title='pytorch_result')
    url = 'http://' + args.inference_addr + '/predictions/' + args.model_name
    tmp_res = requests.post(url, open(args.img, 'rb'))
    server_result = parse_result(tmp_res.json(), model)
    show_result_pyplot(
        model,
        args.img,
        server_result,
        score_thr=args.score_thr,
        title='server_result')
    assert result == server_result


if __name__ == '__main__':
    args = parse_args()
    main(args)
