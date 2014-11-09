// Global variables
BOARD_ELEM = document.getElementById('board')
S = 11;

(function() {
	var selected = null
	var highlighted = []
	var cells = []
	var listeners = []
	BOARD_ELEM.addFieldClickListener = function(listener) {
		listeners.push(listener)
	}
	BOARD_ELEM.updateField = function(row, col, player) {
		var elem = cells[row][col]
		elem.classList[player == +1 ? 'add' : 'remove']('white')
		elem.classList[player == -1 ? 'add' : 'remove']('black')
	}
	BOARD_ELEM.setSelected = function(row, col) {
		BOARD_ELEM.clearSelected()
		cells[row][col].classList.add('selected')
		selected = [row,col]
	}
	BOARD_ELEM.isSelected = function(row, col) {
		return cells[row][col].classList.contains('selected')
	}
	BOARD_ELEM.clearSelected = function() {
		if (selected !== null) {
			cells[selected[0]][selected[1]].classList.remove('selected')
		}
		selected = null
	}
	BOARD_ELEM.getSelected = function() {
		return selected ? selected.slice() : null
	}
	BOARD_ELEM.clearHighlighted = function(fields) {
		var cell
		while ((cell = highlighted.pop())) {
			cell.classList.remove('highlighted')
		}
	}
	BOARD_ELEM.addHighlighted = function(row, col) {
		var cell = cells[row][col]
		cell.classList.add('highlighted')
		highlighted.push(cell)
	}

	var onFieldClicked = function() {
		var parts = this.id.split('_')
		var row = parseInt(parts[1])
		var col = parseInt(parts[2])
		for (var i in listeners) listeners[i](row, col)
	}

	var addLabel = function(text) {
		var label = document.createElement('div')
		row.appendChild(label)
		label.className = 'label'
		label.appendChild(document.createTextNode(text))
	}

	for (var r = 0; r < S; ++r) {
		var row = BOARD_ELEM.appendChild(document.createElement('div'))
		row.className = 'row'
		row.id = 'row_' + r
		addLabel(11 - r)
		cells.push([])
		for (var c = 0; c < S; ++c) {
			var cell = row.appendChild(document.createElement('div'))
			cell.className = 'cell'
			cell.id = 'cell_' + r + '_' + c
			cell.onclick = onFieldClicked
			cells[r].push(cell)
		}
	}
	var row = BOARD_ELEM.appendChild(document.createElement('div'))
	addLabel("")
	for (var c = 0; c < S; ++c) {
		addLabel(String.fromCharCode("A".charCodeAt(0) + c))
	}
})()
