import { createContext } from 'preact';

const NodeStore = createContext([]);
export { NodeStore };

const LocalStore = createContext({
	val: [],
});
export { LocalStore };

const RevStore = createContext({
	val: [],
});
export { RevStore };
