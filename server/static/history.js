HISTORY_ELEM = document.getElementById('history');

(function() {
	'use strict'
	var move_elems = []
	var selected = -1

	function formatCoords(coords) {
		return String.fromCharCode(65 + coords[1]) + (S - coords[0])
	}
	function formatMove(move) {
		return formatCoords(move[0]) + '-' + formatCoords(move[1])
	}

	HISTORY_ELEM.reset = function(moves) {
		for (var i = 0; i < move_elems.length; i += 2) {
			HISTORY_ELEM.removeChild(move_elems[i].parentNode)
		}
		var delim = HISTORY_ELEM.firstChild
		move_elems = []
		selected = -1
		for (var i = 0; i < moves.length; ++i) {
			if (i%2 == 0) {
				var td = document.createElement('td')
				td.className = 'label'
				td.appendChild(document.createTextNode((i/2 + 1) + '.'))
				var tr = document.createElement('tr')
				tr.className = 'move'
				tr.appendChild(td)
				HISTORY_ELEM.insertBefore(tr, delim)
			} else {
				var tr = delim.previousSibling
			}
			var td = document.createElement('td')
			td.className = 'move'
			td.id = 'move_' + i
			td.addEventListener('click', function(event) {
				event.stopPropagation()
				event = document.createEvent("CustomEvent")
				event.initCustomEvent('move-click', true, true, {
					'index': parseInt(this.id.split('_')[1]) })
				HISTORY_ELEM.dispatchEvent(event)
			})
			td.appendChild(document.createTextNode(formatMove(moves[i])))
			tr.appendChild(td)
			move_elems.push(td)
		}

		var holder = document.getElementById('historyHolder')
		holder.scrollTop = holder.scrollHeight
	}

	HISTORY_ELEM.setSelected = function(index) {
		for (var i = 0; i < move_elems.length; ++i) {
			move_elems[i].classList.remove('selected')
			move_elems[i].classList.remove('future')
		}
		selected = index
		if (selected >= 0) {
			move_elems[selected].classList.add('selected')
		}
		for (var i = selected + 1; i < move_elems.length; ++i) {
			move_elems[i].classList.add('future')
		}
	}
	HISTORY_ELEM.getSelected = function() { return selected }
})()
