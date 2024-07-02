import { enqueueRender } from './component';

export let i = 0;

export function createContext(defaultValue) {
	const ctx = {};

	const context = {
		_id: '__cC' + i++,
		_defaultValue: defaultValue,
		Consumer(props, context) {
			return props.children(context);
		},
		Provider(props) {
			if (!this.getChildContext) {
				const subs = [];
				this.shouldComponentUpdate = _props => {
					if (this.props.value !== _props.value) {
							payload.c.context = _props.value;
							if (!payload.u || payload.u(_props.value, this.props.value)) {
								enqueueRender(payload.c);
				this.sub = (c, shouldUpdate) => {
					const entry = { c, u: shouldUpdate };
					subs.push(entry);
					let old = c.componentWillUnmount;
					c.componentWillUnmount = () => {
						subs.splice(subs.indexOf(entry), 1);
						old && old.call(c);
					};
				this.getChildContext = () => {
					ctx[context._id] = this;
					return ctx;
				};

				this.shouldComponentUpdate = function (_props) {
					const newValue = this.props.value;
					if (newValue !== _props.value) {
						subs.some(payload => {
							payload[0].context = _props.value;
							if (!payload[1] || payload[1](_props.value, newValue)) {
								enqueueRender(payload[0]);
							}
						});
					}
				};

				this.sub = function() {
					subs.push(arguments);
				};
			}

			return props.children;
		}
	};

	context.Consumer.contextType = context;

	return context;
}
