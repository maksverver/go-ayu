(function(){
	'use strict'
	var state = null
	var my_last_version = null

	var getPlayerKey = function(player) {
		if (player == +1) return getParameter('white')
		if (player == -1) return getParameter('black')
		return null
	}

	BOARD_ELEM.addEventListener('field-click', function(event) {
		if (!state || !getPlayerKey(state.nextPlayer)) return
		if (HISTORY_ELEM.getSelected() != state.history.length - 1) {
			selectMove(state.history.length - 1)
		}
		var row = event.detail.row
		var col = event.detail.col
		var player = state.fields[row][col]
		if (player == state.nextPlayer) {
			if (BOARD_ELEM.isSelected(row, col)) {
				BOARD_ELEM.clearSelected()
			} else {
				BOARD_ELEM.setSelected(row, col)
			}
		} else if (player == 0 && BOARD_ELEM.getSelected()) {
			var req = new XMLHttpRequest()
			req.onreadystatechange = function(){
				if (req.readyState == 4) {
					if (req.status != 200) {
						alert("Update request failed!\n" + req.responseText)
					}
				}
			}
			req.open('POST', 'update', true)
			req.setRequestHeader("Content-type", "application/json")
			req.send(JSON.stringify({
				'game': getParameter('game'),
				'version': state.history.length,
				'key': getPlayerKey(state.nextPlayer),
				'move': [BOARD_ELEM.getSelected(), [row,col]]}))
			BOARD_ELEM.clearSelected()
			my_last_version = state.history.length + 1
		}
	})

	HISTORY_ELEM.addEventListener('move-click', function(event) {
		selectMove(event.detail.index)
	})

	var selectMove = function(index) {
		var fields = JSON.parse(JSON.stringify(state.fields))
		for (var i = state.history.length - 1; i > index; --i) {
			var move = state.history[i]
			var tmp = fields[move[0][0]][move[0][1]]
			fields[move[0][0]][move[0][1]] = fields[move[1][0]][move[1][1]]
			fields[move[1][0]][move[1][1]] = tmp
		}
		BOARD_ELEM.clearSelected()
		BOARD_ELEM.clearHighlighted()
		BOARD_ELEM.setFields(fields)
		if (index >= 0) {
			var move = state.history[index]
			BOARD_ELEM.addHighlighted(move[0][0], move[0][1])
			BOARD_ELEM.addHighlighted(move[1][0], move[1][1])
			HISTORY_ELEM.setSelected(index)
		}
	}

	var update = function() {  // called whenever the game state changes
		BOARD_ELEM.clearSelected()
		BOARD_ELEM.setFields(state.fields)
		document.getElementById('whiteToMove').style.display =
			(state.nextPlayer == +1) ? '' : 'none'
		document.getElementById('blackToMove').style.display =
			(state.nextPlayer == -1) ? '' : 'none'
		document.getElementById('yourTurn').style.display =
			getPlayerKey(state.nextPlayer) ? '' : 'none'

		HISTORY_ELEM.reset(state.history)
		selectMove(state.history.length - 1)

		if (my_last_version === null) {
			my_last_version = state.history.length
		} else if (my_last_version != state.history.length &&
			document.getElementById('playSound').checked) {
			document.getElementById('turnNotification').play()
		}

		CLOCK.setPlayer(state.nextPlayer)
		if (state.timeUsed) {
			CLOCK.setTimeUsed(state.timeUsed[0], state.timeUsed[1])
		}
	}

	var pollState = function(game, version) {
		var pollDelay = 10e3  // 10 seconds (in milliseconds)
		var startTime = new Date().getTime()
		var req = new XMLHttpRequest()
		req.onreadystatechange = function(){
			if (req.readyState == 4) {
				if (req.status >= 400) {
					alert("Poll request failed!\n" + req.responseText + "\nYou probably need to refresh the page.")
					return
				}
				if (req.responseText) {
					state = JSON.parse(req.responseText)
					update()
				}
				var new_version = state.history.length + 1
				var repoll = function() {
					pollState(game, new_version)
				}
				var delay = new Date().getTime() - startTime
				if (delay > pollDelay || new_version != version) {
					repoll()
				} else {
					setTimeout(repoll, pollDelay - delay)
				}
			}
		}
		req.open('GET', 'poll?game=' + encodeURIComponent(game) + '&version=' + version, true)
		req.send()
	}

	pollState(getParameter('game'), 0)

})()
