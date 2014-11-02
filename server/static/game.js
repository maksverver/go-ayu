(function(){
	var state = null
	var my_last_version = null

	var getPlayerKey = function(player) {
		if (player == +1) return getParameter('white')
		if (player == -1) return getParameter('black')
		return null
	}

	BOARD_ELEM.addFieldClickListener(function(row,col) {
		if (!state || !getPlayerKey(state.nextPlayer)) return
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
			req.send(JSON.stringify({
				'game': getParameter('game'),
				'version': state.history.length,
				'key': getPlayerKey(state.nextPlayer),
				'move': [BOARD_ELEM.getSelected(), [row,col]]}))
			BOARD_ELEM.clearSelected()
			my_last_version = state.history.length + 1
		}
	})

	var updateBoard = function() {
		BOARD_ELEM.clearSelected()
		for (var i = 0; i < state.size; ++i) {
			for (var j = 0; j < state.size; ++j) {
				BOARD_ELEM.updateField(i, j, state.fields[i][j])
			}
		}
		document.getElementById('whiteToMove').style.display =
			(state.nextPlayer == +1) ? '' : 'none'
		document.getElementById('blackToMove').style.display =
			(state.nextPlayer == -1) ? '' : 'none'
		document.getElementById('yourTurn').style.display =
			getPlayerKey(state.nextPlayer) ? '' : 'none'

		if (my_last_version === null) {
			my_last_version = state.history.length
		} else if (my_last_version != state.history.length &&
			document.getElementById('playSound').checked) {
			document.getElementById('turnNotification').play()
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
					updateBoard()
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
