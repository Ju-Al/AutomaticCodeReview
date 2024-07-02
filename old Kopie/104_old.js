import { Component, createElement } from 'react';

import classify from 'src/classify';
import ResetButton from './resetButton';
import defaultClasses from './exit.css';

class Exit extends Component {
    static propTypes = {
        classes: shape({
            body: string,
            footer: string,
            root: string
        }),
        resetCheckout: func
    };

    render() {
        const { classes, resetCheckout } = this.props;

        return (
            <div className={classes.root}>
                <div className={classes.body}>Thank you for your order!</div>
                <div className={classes.footer}>
                    <ResetButton resetCheckout={resetCheckout} />
                </div>
            </div>
        );
    }
}

export default classify(defaultClasses)(Exit);
