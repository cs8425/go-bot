import { h, Fragment } from 'preact';
import { useState, useEffect, useRef, useImperativeHandle } from 'preact/hooks';
import { forwardRef } from 'preact/compat';

import BackupIcon from '@material-ui/icons/Backup';

import { makeStyles } from "@material-ui/core/styles";

const useStyles = makeStyles((theme) => ({
	zoneDrag: {
		'& *': {
			'pointer-events': 'none',
		},
	},
	zoneHint: {
		'z-index': 10,
		'border-style': 'dashed',
		'padding': '1rem',
		'background-color': theme.palette.warning.light,
		'& *': {
			'pointer-events': 'none',
		},
	},
}));

const DragNdrop = forwardRef((props, ref) => {
	const { children, handleFile, accept, onClick, ...other } = props;
	const classes = useStyles();
	const inputEl = useRef(null);
	const zoneEl = useRef(null);
	const [zoneEnter, setZoneEnter] = useState(false);

	useEffect(() => {
		return () => {
			if (inputEl.current) inputEl.current.value = null;
			// console.log('[drag][clean]', inputEl);
		};
	}, [inputEl]);

	const handleFiles = (e) => {
		let files = e?.target?.files || e?.dataTransfer?.files;
		// console.log('[drag]handleFiles', this, e, files);
		handleFile(files);
	}
	const drop = (e) => {
		e.preventDefault();
		// console.log('[drag]drop', this, e, inputEl);
		setZoneEnter(false);
		handleFiles(e);
	}
	const isInZone = (e) => {
		let ele = e.target;
		if (ele == zoneEl.current) return true;
		let i = 0;
		while (ele) {
			if (ele.parentNode == zoneEl.current) return true;
			// console.log('[zone]', i, ele.parentNode, zoneEl.current)
			ele = ele.parentNode;
			i++;
		}
		return false;
	}
	let handleClick = (e) => inputEl.current.click();
	if (onClick === false) {
		handleClick = null;
	} else {
		if (typeof onClick === 'function') handleClick = (e) => onClick(e);
	}
	useImperativeHandle(ref, () => ({
		open: () => {
			inputEl.current.click();
		},
	}));

	return (
		<div
			ref={zoneEl}
			ondragenter={(e) => setZoneEnter(isInZone(e))}
			ondragleave={(e) => isInZone(e) && setZoneEnter(e.target != zoneEl.current)}
			ondrop={drop}
			ondragover={(e) => e.preventDefault()}
			onClick={handleClick}
			style="position: relative;"
			className={zoneEnter ? classes.zoneDrag : null}
		>
			{children}
			<input ref={inputEl} accept={accept || ''} type="file" autocomplete="off" tabindex="-1" style="display: none;" onchange={handleFiles}></input>
			{zoneEnter &&
				<div className={zoneEnter ? classes.zoneHint : null} style="position: absolute; text-align: center; left: 0; right: 0; top: 0; bottom: 0;">
					<BackupIcon style={{ fontSize: '5rem' }} ></BackupIcon>
				</div>
			}
		</div>
	);
});

export default DragNdrop;
export { DragNdrop };
