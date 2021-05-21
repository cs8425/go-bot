import { h, Component } from 'preact';
import { useRoute } from 'wouter-preact';

// class style
class KeepStateRoute extends Component {
	render(props, state) {
		const [isActive, params] = useRoute(props.path);
		return (isActive) ? h('div', null, props.children) : h('div', { style: 'display: none' }, props.children);
	}
}

// hook style
// const KeepStateRoute = (props) => {
// 	const { path, match, component, children } = props;
// 	const useRouteMatch = useRoute(path);
// 	const [matches, params] = match || useRouteMatch;
// 	return (matches) ? h('div', null, children) : h('div', { style: 'display: none' }, children);
// };

export { KeepStateRoute };
