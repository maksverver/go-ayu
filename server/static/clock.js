CLOCK = {};

(function() {
	'use strict'
	var player = 0
	var white_millis = 0
	var black_millis = 0
	var cur_time_millis = new Date().getTime()

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
		var new_time_millis = new Date().getTime()
		var millis = new_time_millis - cur_time_millis
		cur_time_millis = new_time_millis
		if (player == +1) white_millis += millis
		if (player == -1) black_millis += millis
		updateTimeText('whiteTime', white_millis)
		updateTimeText('blackTime', black_millis)
	}

	setInterval(updateTime, 1000)

	CLOCK.setPlayer = function(p) {
		updateTime()
		player = p
	}
	CLOCK.setTimeUsed = function(white_seconds, black_seconds) {
		white_millis = parseInt(1000*white_seconds)
		black_millis = parseInt(1000*black_seconds)
		updateTime()
	}
})()
