// Global variables
BOARD_ELEM = document.getElementById('board')
BOARD_SIZE = getParameter('size') || 11;

(function() {
	'use strict'
	var selected = null
	var highlighted = []
	var cells = []
	BOARD_ELEM.updateField = function(row, col, player) {
		var elem = cells[row][col]
		elem.classList[player == +1 ? 'add' : 'remove']('white')
		elem.classList[player == -1 ? 'add' : 'remove']('black')
	}
	BOARD_ELEM.setFields = function(fields) {
		for (var i = 0; i < fields.length; ++i) {
			for (var j = 0; j < fields[i].length; ++j) {
				BOARD_ELEM.updateField(i, j, fields[i][j])
			}
		}
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

	var addLabel = function(text) {
		var label = document.createElement('div')
		row.appendChild(label)
		label.className = 'label'
		label.appendChild(document.createTextNode(text))
	}


	for (var r = BOARD_SIZE - 1; r >= 0; --r) {
		var row = BOARD_ELEM.appendChild(document.createElement('div'))
		row.className = 'row'
		row.id = 'row_' + r
		addLabel(r + 1)
		cells[r] = []
		for (var c = 0; c < BOARD_SIZE; ++c) {
			var cell = row.appendChild(document.createElement('div'))
			cell.className = 'cell'
			cell.id = 'cell_' + r + '_' + c
			cell.addEventListener('click', function(event) {
				event.stopPropagation()
				var parts = this.id.split('_')
				event = document.createEvent("CustomEvent")
				event.initCustomEvent('field-click', true, true, {
					'row': parseInt(parts[1]),
					'col': parseInt(parts[2]) })
				BOARD_ELEM.dispatchEvent(event)
			})
			cells[r][c] = cell
		}
	}
	var row = BOARD_ELEM.appendChild(document.createElement('div'))
	addLabel("")
	for (var c = 0; c < BOARD_SIZE; ++c) {
		addLabel(String.fromCharCode("A".charCodeAt(0) + c))
	}
})()
