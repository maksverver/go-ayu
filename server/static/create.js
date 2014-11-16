(function(){
	'use strict'

	var showGameLinks = function(game_id, keys, size) {
		document.getElementById('loading').style.display = 'none'
		document.getElementById('loaded').style.display = ''
		for (var i = 0; i < 4; ++i) {
			var a = document.getElementById('link-' + i)
			a.target = 'game-' + game_id + '_' + i
			a.href = 'game.html#game=' + encodeURIComponent(game_id)
			if (i&1) a.href += '&white=' + encodeURIComponent(keys[0])
			if (i&2) a.href += '&black=' + encodeURIComponent(keys[1])
			a.href += '&size=' + size
		}
	}

	var createGame = function(size) {
		console.log("Creating game with board size " + size)
		var req = new XMLHttpRequest()
		req.onreadystatechange = function(){
			if (req.readyState == 4) {
				if (req.status != 200) {
					alert("Create request failed!\n" + req.responseText)
					return
				}
				var res = JSON.parse(req.responseText)
				showGameLinks(res.game, res.keys, res.size)
			}
		}
		req.open('POST', 'create', true)
		req.send(JSON.stringify({"size": size}))
	}

	document.getElementById('createGameForm').onsubmit = function() {
		createGame(parseInt(document.getElementById('boardSize').value))
		return false
	}
})()
