    print('-'*88)

    if args.action == 'deploy_database':
        print("Deploying the serverless database cluster and supporting resources.")
        do_deploy_database(cluster_name, secret_name)
        print("Next, run 'py library_demo.py deploy_rest' to deploy the REST API.")
    elif args.action == 'deploy_rest':
        print("Deploying the REST API components.")
        api_url = do_deploy_rest(stack_name)
        print(f"Next, send HTTP requests to {api_url} or run "
              f"'py library_demo.py demo_rest' "
              f"to see a demonstration of how to call the REST API by using the "
              f"Requests package.")
    elif args.action == 'demo_rest':
        print("Demonstrating how to call the REST API by using the Requests package.")
        try:
            do_rest_demo(stack_name)
        except TimeoutError as err:
            print(err)
        else:
            print("Next, give it a try yourself or run 'py library_demo.py cleanup' "
                  "to delete all demo resources.")
    elif args.action == 'cleanup':
        print("Cleaning up all resources created for the demo.")
        do_cleanup(stack_name, cluster_name, secret_name)
        print("All clean, thanks for watching!")
    print('-'*88)


if __name__ == '__main__':
    main()
