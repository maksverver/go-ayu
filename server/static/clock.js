CLOCK = {};

(function() {
	var player = 0
	var whiteMillis = 0
	var blackMillis = 0
	var curTimeMillis = new Date().getTime()

	var updateTimeText = function(elem_id, millis) {
		var seconds = parseInt(millis/1000)
		var minutes = parseInt(seconds/60)
		seconds %= 60
		if (seconds < 10) seconds = '0' + seconds
		var elem = document.getElementById(elem_id)
		while (elem.firstChild) elem.removeChild(elem.firstChild)
		elem.appendChild(document.createTextNode(minutes + ':' + seconds))
	}

	var updateTime = function() {
		var newTimeMillis = new Date().getTime()
		var millis = newTimeMillis - curTimeMillis
		curTimeMillis = newTimeMillis
		if (player == +1) whiteMillis += millis
		if (player == -1) blackMillis += millis
		updateTimeText('whiteTime', whiteMillis)
		updateTimeText('blackTime', blackMillis)
	}

	setInterval(updateTime, 1000)

	CLOCK.setPlayer = function(p) {
		updateTime()
		player = p
	}
	CLOCK.setTimeUsed = function(whiteSeconds, blackSeconds) {
		whiteMillis = parseInt(1000*whiteSeconds)
		blackMillis = parseInt(1000*blackSeconds)
		updateTime()
	}
})()
