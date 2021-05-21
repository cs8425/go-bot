import { h, Fragment, Component, render } from 'preact';
import { useState, useEffect, useContext } from 'preact/hooks';

class Counter extends Component {
	state = {
		value: 0,
	};
	setValue(v) {
		this.setState({ value: v });
	}
	render({ }, { value }) {
		return (
			<>
				<div>Counter: {value}</div>
				<button onClick={() => this.setValue(value + 1)}>Increment</button>
				<button onClick={() => this.setValue(value - 1)}>Decrement</button>
			</>
		);
	}
}

export { Counter };


import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TablePagination from '@material-ui/core/TablePagination';
import TableRow from '@material-ui/core/TableRow';

function DataList(props) {
	const { children, data, header, renderRow, ...other } = props;
	const [page, setPage] = useState(0);
	const [rowsPerPage, setRowsPerPage] = useState(10);

	const handleChangePage = (event, newPage) => {
		setPage(newPage);
	};

	const handleChangeRowsPerPage = (event) => {
		setRowsPerPage(+event.target.value);
		setPage(0);
	};

	let rows = data || [];
	return (
		<>
			<TableContainer>
				<Table stickyHeader aria-label="sticky table">
					<TableHead>{header}</TableHead>
					<TableBody>
						{rows.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((row, idx) => {
							return renderRow(row, idx);
						})}
					</TableBody>
				</Table>
			</TableContainer>
			<TablePagination
				rowsPerPageOptions={[10, 25, 50, 100]}
				component="div"
				count={rows.length}
				rowsPerPage={rowsPerPage}
				page={page}
				onChangePage={handleChangePage}
				onChangeRowsPerPage={handleChangeRowsPerPage}
			/>
		</>
	);
}

export { DataList };
