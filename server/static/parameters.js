(function(){
	var parameters = null
	
	parseParameters = function() {
		parameters = {}
		var parts = window.location.hash.replace(/^#/,'').split('&')
		for (var i = 0; i < parts.length; ++i) {
			var keyval = parts[i].split('=')
			if (keyval.length == 2) {
				var key = decodeURIComponent(keyval[0])
				var val = decodeURIComponent(keyval[1])
				parameters[key] = val
			}
		}
	}
	
	var updateParameters = function() {
		var s = "#"
		for (var key in parameters) {
			if (s != "#") s += '&'
			s += encodeURIComponent(key) + '=' +
				encodeURIComponent(parameters[key])
		}
		window.location.hash = s
	}
	
	getParameter = function(key, def) {
		if (parameters === null) parseParameters()
		var res = parameters[key]
		return (res === undefined) ? def : res
	}
	
	setParameter = function(key, val) {
		if (parameters === null) parseParameters()
		if (val === undefined) {
			delete parameters[key]
		} else {
			parameters[key] = String(val)
		}
		updateParameters()
	}
})()
