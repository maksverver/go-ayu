(function(){
	'use strict'
	var req = new XMLHttpRequest()
	req.onreadystatechange = function(){
		if (req.readyState == 4) {
			if (req.status != 200) {
				alert("Create request failed!\n" + req.responseText)
				return
			}
			document.getElementById('loading').style.display = 'none'
			document.getElementById('loaded').style.display = ''
			var res = JSON.parse(req.responseText)
			for (var i = 0; i < 4; ++i) {
				var a = document.getElementById('link-' + i)
				a.target = 'game-' + res.game + '_' + i
				a.href = 'game.html#game=' + encodeURIComponent(res.game)
				if (i&1) a.href += '&white=' + encodeURIComponent(res.keys[0])
				if (i&2) a.href += '&black=' + encodeURIComponent(res.keys[1])
			}
		}
	}
	req.open('POST', 'create', true)
	req.send()
})()
