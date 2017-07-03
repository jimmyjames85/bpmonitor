var SESSION_ID_COOKIE_KEY = "eWVrc2loV2hzYU1ydW9TZWVzc2VubmVUeXRpbGF1UWRuYXJCNy5vTmRsT2VtaXRkbE9zJ2xlaW5hRGtjYUoK";

function getElement(name){
    return document.getElementById(name)
}

function getCookie(cname)
{
	var name = cname + "=";
	var ca = document.cookie.split(';');
	for (var i = 0; i < ca.length; i++)
	{
		var c = ca[i];
		while (c.charAt(0) == ' ')
			// skip spaces?
			c = c.substring(1);
		if (c.indexOf(name) != -1)
			return c.substring(name.length, c.length);
	}
	return "";
}

function objectString(obj)
{
	var ret = "";
	for ( var key in obj)
	    ret += key + ": " + obj[key] + ";\n\n ";
	return ret;
}

function describeObject(obj)
{
	var ret = "";
	for ( var key in obj)
		ret += "<b>" + key + "</b> = " + obj[key] + "<br>";
	return ret;
}

/**
 * Sets a cookie for the page and path If path is '/' then a global cookie is
 * set
 * 
 * @param cname -
 *            name of the cookie
 * @param cvalue -
 *            value to paired with cname
 * @param exdays -
 *            expiration date (client side)
 * @param path -
 *            path relative base url
 * 
 * 
 */
function setCookie(cname, cvalue, exdays, path)
{
	var d = new Date();
	d.setTime(d.getTime() + (exdays * 24 * 60 * 60 * 1000));
	var expires = "expires=" + d.toUTCString();
	document.cookie = cname + "=" + cvalue + "; " + expires + "; path=" + path;
}

/**
 * 
 * @param obj
 *            the object from which to pull data from for every member of obj
 *            (e.g. obj.key1) it assigns the value of obj.key1 to the member to
 *            the key
 * 
 * @returns {String} returns key1=val1&key2=val2...
 */
function createUrlEncodedRequest(obj)
{
	var request = "";
	var count = 0;
	for ( var key in obj)
	{
		var k = encodeURIComponent(key);
		var v = encodeURIComponent(obj[key]);

		if (count == 0)
			request += k + "=" + v;
		else
			request += "&" + k + "=" + v;
		count++;
	}
	return request;
}

function getBaseURL()
{
    var url = window.location.href; // entire url including querystring
    
    end = url.indexOf('/', 7); // 7 = len('http://')
    if (url.substring(0, 5)=='https') 
	end = url.indexOf('/', 8);

    // if we don't want to include the port
    port = url.indexOf(':', 6); // 6 avoids the first :
    if ( port >= 0 && port < end) {
    	end = port;
    }

    var baseURL = url.substring(0, end)
    return baseURL;
}

function sendAjaxQuery(method, url, urlencodedRequest, callback)
{
	console.debug("old method");
	var emptyHeaders = new Object();
	sendAjaxQueryWithHeaders(method, url, urlencodedRequest, emptyHeaders, callback);
}

function sendAjaxQueryWithHeaders(method, url, urlencodedRequest, headers, callback)
{
	// https://developer.mozilla.org/en-US/docs/Web/API/XMLHttpRequest
	var xmlhttp;

	if (window.XMLHttpRequest)
	{// code for IE7+, Firefox, Chrome, Opera, Safari
		xmlhttp = new XMLHttpRequest();
	}
	else
	{// code for IE6, IE5
		xmlhttp = new ActiveXObject("Microsoft.XMLHTTP");
	}
	xmlhttp.onreadystatechange = function()
	{
		if (xmlhttp.readyState == 4 && xmlhttp.status == 200)
		{
			if (callback)
				callback(xmlhttp.responseText);
		}
	}
	xmlhttp.open(method, url, true);
	xmlhttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");

	for ( var key in headers)
		xmlhttp.setRequestHeader(key, headers[key]);

	// xmlhttp.setRequestHeader("getBoard","true");//TODO send custom headers as
	// a parameter in the function
	xmlhttp.send(urlencodedRequest);
}

function sendAjaxQueryJsonObject(method, url, urlencodedRequest, callback)
{
	sendAjaxQuery(method, url, urlencodedRequest, function(json)
	{
		callback(JSON.parse(json));
	});
}

function setInnerHtml(id, html)
{
	document.getElementById(id).innerHTML = html;
}
