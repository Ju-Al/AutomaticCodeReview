def handle_error_by_throwing_exception():


def handle_error_by_returning_none(input_data):
    try:
        return int(input_data)
    except ValueError:
        return None


def handle_error_by_returning_tuple(input_data):
    try:
        return (True, int(input_data))
    except ValueError:
        return (False, None)


def filelike_objects_are_closed_on_exception(filelike_object):
    with filelike_object:
        filelike_object.close()
        raise Exception()
