//cookies
// var SESSION_ID_COOKIE_KEY = "s";
var SESSION_ID_COOKIE_KEY = "eWVrc2loV2hzYU1ydW9TZWVzc2VubmVUeXRpbGF1UWRuYXJCNy5vTmRsT2VtaXRkbE9zJ2xlaW5hRGtjYUoK";

// servlet information for getting user objects
var LOCAL_HOST_PROJ_NAME = "BoColyer";
var USERS_SERVLET = "/usersServlet";


var GAME_SERVLET = "/DiplomaCY/game";
var GAME_MANAGER = "/DiplomaCY/RunManager";
var MOVES_SERVLET = "/DiplomaCY/moves";

var ACTION_GET_VIA_SESSIONID = "GETS";
var ACTION_KEY = "action";
var SID_KEY = "sid";


function getElement(name){
    return document.getElementById(name)
}

function doesUserExist(username, callback)
{
	var request = new Object();
	request.username = username;
	request.action = "PINGUSER";
	var urlencodedRequest = createUrlEncodedRequest(request);
	sendAjaxQuery("POST", getBaseURL() +  USERS_SERVLET, urlencodedRequest, callback)
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
 * A global session cookie is set whenever userLogin() is called successfully.
 * If the ajax call to retrieve the user by session id is successful then this
 * method passes the response to callback(response)
 * 
 * @param callback
 */
function getCurrentUser(callback)
{
	var sessionId = getCookie(SESSION_ID_COOKIE_KEY);
	var request = new Object();
	request[ACTION_KEY] = ACTION_GET_VIA_SESSIONID;
	request[SID_KEY] = sessionId;
	sendAjaxQuery("POST", getBaseURL() + USERS_SERVLET, createUrlEncodedRequest(request), callback);
}

function createUser(username, password, email, firstName, lastName, callback)
{
	var request = new Object();
	request.username = username;
	request.password = password;
	request.email = email;
	request.firstName = firstName;
	request.lastName = lastName;
	request.action = "ADD";

	
	sendAjaxQuery("POST", getBaseURL() + USERS_SERVLET, createUrlEncodedRequest(request), callback)
	
}

function _userLogin(username, password, callBack)
{
	var request = new Object();
	request.username = username;
	request.password = password;
	request.action = "LOGIN";
	sendAjaxQuery("POST", getBaseURL() + USERS_SERVLET, createUrlEncodedRequest(request), callBack);
}

/**
 * 
 * if a user is logged in returns a json of the user, and then logs the user
 * out. else returns a message saying no user logged in
 * 
 * @param callBack-return
 *            callback method
 */
function userLogout(callBack)
{
	var sid = getCookie("s");

	var request = new Object();
	request.sid = sid;
	request.action = "LOGOUT";

	sendAjaxQuery("POST", getBaseURL() + USERS_SERVLET, createUrlEncodedRequest(request), callBack);

	// delete cookie
	setCookie(SESSION_ID_COOKIE_KEY, "", -1, "/");

}

/**
 * 
 * @param user
 *            -should be an object properties user.username user.password and
 *            any changed properties
 * 
 * @param callBack -
 *            the method to call when the server responds
 */
function updateUserInfo(user, callBack)
{
	// the user should have all the query elements already (i.e. username,
	// password, lastName etc) So we just set the action element
	user.action = "UPDATE";
	sendAjaxQuery("POST", getBaseURL() + USERS_SERVLET, createUrlEncodedRequest(user), callBack);
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

function getGameServlet()
{
	return getBaseURL() + GAME_SERVLET;
}

function getGameMap(gameId, callback)
{
	var sessionId = getCookie(SESSION_ID_COOKIE_KEY);

	var headers = new Object();
	headers.getBoard = true;
	headers.gameId = gameId;
	headers.sessionId = sessionId;

	sendAjaxQueryWithHeaders("GET", getBaseURL() + GAME_SERVLET, "", headers, callback);
}

function triggerClock(gameId, callback)
{
	var headers = new Object();
	headers.gameId = gameId;

	sendAjaxQueryWithHeaders("GET", getBaseURL() + GAME_MANAGER, "", headers, callback);
}

function submitMoves(gameId, movesArr, callback)
{
	var sessionId = getCookie(SESSION_ID_COOKIE_KEY);

	var headers = new Object();
	headers.postMoves = true;
	headers.gameId = gameId;
	headers.sessionId = sessionId;

	sendAjaxQueryWithHeaders("POST", getBaseURL() + MOVES_SERVLET, JSON.stringify(movesArr), headers, callback);
}

function getCurrentUsersGames(callback)
{
	var sessionId = getCookie(SESSION_ID_COOKIE_KEY);

	var parameters = new Object();
	parameters.sid = sessionId;
	parameters.action = "GETGAMES";

	sendAjaxQuery("POST", getBaseURL() + USERS_SERVLET, createUrlEncodedRequest(parameters), callback);
}

function getCurrentGameState(gameId, callback)
{
	var sessionId = getCookie(SESSION_ID_COOKIE_KEY);
	var headers = new Object();
	headers.sessionId = sessionId;
	headers.gameId = gameId;
	headers.getCurrentState = true;
	sendAjaxQueryWithHeaders("GET", getBaseURL() + GAME_SERVLET, "", headers, callback);
}

function getSubmitStatus(gameId, callback)
{
	var sessionId = getCookie(SESSION_ID_COOKIE_KEY);
	var headers = new Object();
	headers.sessionId = sessionId;
	headers.gameId = gameId;
	headers.getSubmitStatus= true;
	sendAjaxQueryWithHeaders("GET", getBaseURL() + MOVES_SERVLET, "", headers, callback);
	
}

function getPendingMoves(gameId, callback)
{
	var sessionId = getCookie(SESSION_ID_COOKIE_KEY);
	var headers = new Object();
	headers.sessionId = sessionId;
	headers.gameId = gameId;
	headers.getPendingMoves= true;
	sendAjaxQueryWithHeaders("GET", getBaseURL() + MOVES_SERVLET, "", headers, callback);
}

function getUserList(gameId, callback)
{
	var sessionId = getCookie(SESSION_ID_COOKIE_KEY);
	var headers = new Object();
	headers.sessionId = sessionId;
	headers.gameId = gameId;
	headers.getUserList= true;
	sendAjaxQueryWithHeaders("GET", getBaseURL() + GAME_SERVLET, "", headers, callback);	
}

function sendAddFriend(friendname, callback)
{
	var sessionId = getCookie(SESSION_ID_COOKIE_KEY);
	
	var headers = new Object();
	headers.sessionId = sessionId;
	headers.newFriend = friendname;
	headers.action = "ADDFRIEND";
	sendAjaxQueryWithHeaders("POST", getBaseURL() + USERS_SERVLET, "", headers, callback);
}

function getLastMoves(gameId, curStateNumber, callback)
{
	var sessionId = getCookie(SESSION_ID_COOKIE_KEY);
	var headers = new Object();
	headers.sessionId = sessionId;
	headers.gameId = gameId;
	headers.stateNumber = curStateNumber -1;	
	headers.getLastMoves= true;
	sendAjaxQueryWithHeaders("GET", getBaseURL() + MOVES_SERVLET, "", headers, callback);	
}

function sendDropFriend(friendname, callback)
{
	var sessionId = getCookie(SESSION_ID_COOKIE_KEY);

	var headers = new Object();
	headers.sessionId = sessionId;
	headers.exFriend = friendname;
	headers.action = "DROPFRIEND";
	sendAjaxQueryWithHeaders("POST", getBaseURL() + USERS_SERVLET, "", headers, callback);
}
