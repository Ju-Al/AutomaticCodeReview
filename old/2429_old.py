    # support batch graph.
    for i in range(len(dataset)):
        dataset.graph_lists[i] = dgl.add_self_loop(dataset.graph_lists[i])

    num_training = int(len(dataset) * 0.8)
    num_val = int(len(dataset) * 0.1)
    num_test = len(dataset) - num_val - num_training
    train_set, val_set, test_set = random_split(dataset, [num_training, num_val, num_test])

    train_loader = GraphDataLoader(train_set, batch_size=args.batch_size, shuffle=True, num_workers=6)
    val_loader = GraphDataLoader(val_set, batch_size=args.batch_size, num_workers=2)
    test_loader = GraphDataLoader(test_set, batch_size=args.batch_size, num_workers=2)

    device = torch.device(args.device)
    
    # Step 2: Create model =================================================================== #
    num_feature, num_classes, _ = dataset.statistics()
    model_op = get_sag_network(args.architecture)
    model = model_op(in_dim=num_feature, hid_dim=args.hid_dim, out_dim=num_classes,
                     num_convs=args.conv_layers, pool_ratio=args.pool_ratio, dropout=args.dropout).to(device)
    args.num_feature = int(num_feature)
    args.num_classes = int(num_classes)

    # Step 3: Create training components ===================================================== #
    optimizer = torch.optim.Adam(model.parameters(), lr=args.lr, weight_decay=args.weight_decay)

    # Step 4: training epoches =============================================================== #
    bad_cound = 0
    best_val_loss = float("inf")
    final_test_acc = 0.
    best_epoch = 0
    train_times = []
    for e in range(args.epochs):
        s_time = time()
        train_loss = train(model, optimizer, train_loader, device)
        train_times.append(time() - s_time)
        val_acc, val_loss = test(model, val_loader, device)
        test_acc, _ = test(model, test_loader, device)
        if best_val_loss > val_loss:
            best_val_loss = val_loss
            final_test_acc = test_acc
            bad_cound = 0
            best_epoch = e + 1
        else:
            bad_cound += 1
        if bad_cound >= args.patience:
            break
        
        if (e + 1) % args.print_every == 0:
            log_format = "Epoch {}: loss={:.4f}, val_acc={:.4f}, final_test_acc={:.4f}"
            print(log_format.format(e + 1, train_loss, val_acc, final_test_acc))
    print("Best Epoch {}, final test acc {:.4f}".format(best_epoch, final_test_acc))
    return final_test_acc, sum(train_times) / len(train_times)


if __name__ == "__main__":
    args = parse_args()
    res = []
    train_times = []
    for i in range(args.num_trials):
        print("Trial {}/{}".format(i + 1, args.num_trials))
        acc, train_time = main(args)
        res.append(acc)
        train_times.append(train_time)

    mean, err_bd = get_stats(res, conf_interval=True)
    print("mean acc: {:.4f}, error bound: {:.4f}".format(mean, err_bd))

    out_dict = {"hyper-parameters": vars(args),
                "result": "{:.4f}(+-{:.4f})".format(mean, err_bd),
                "train_time": "{:.4f}".format(sum(train_times) / len(train_times))}

    with open(args.output_path, "w") as f:
        json.dump(out_dict, f, sort_keys=True, indent=4)
