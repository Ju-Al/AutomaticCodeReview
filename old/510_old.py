

def get_changefeed():
    return ChangeFeed(table='votes', operation='insert')


def create_pipeline():
    election = Election()

    election_pipeline = Pipeline([
        Node(election.check_for_quorum),
        Node(election.requeue_transactions)
    ])

    return election_pipeline


def start():
    pipeline = create_pipeline()
    pipeline.setup(indata=get_changefeed())
    pipeline.start()
    return pipeline
