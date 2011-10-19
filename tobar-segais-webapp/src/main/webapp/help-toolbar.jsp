<html lang="en">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">

<title>Search:</title>
     
<style type="text/css">
/* need this one for Mozilla */
html {
	width:100%;
	height:100%;
	margin:0;
	padding:0;
	border:0;
 }

body {
	background:ButtonFace;
	border:0;
	height:100%;
}

table {
	font: icon;
	background:ButtonFace;
	margin: 0;
	padding: 0;
	height:100%;
}

form {
	background:ButtonFace;
	height:100%;
	margin:0;
}

input {
	font: icon;
	margin:0;
	padding:0;
}

input {
    font-size: 1.0em;
}

a {
	color:WindowText;
	text-decoration:none;
}

#searchTD {
	padding-left:7px;
	padding-right:4px;
}

#searchWord {
	margin-left:5px;
	margin-right:5px;
	border:1px solid ThreeDShadow;
}

#searchLabel {
	color:WindowText;
}

#go {
    background:GrayText;
	color:Window;
	font-weight:bold;
	border:1px solid ThreeDShadow;
	margin-left:1px;
	font-size: 1.0em;
}

#scopeLabel {
	text-decoration:underline; 
	color:#0066FF; 
	cursor:pointer;
	padding-left:15px;   /* This should be the same for both RTL and LTR. */
}

#scope { 
	text-align:right;
	margin-left:5px;
	border:0;
	color:WindowText;
	text-decoration:none;
}

</style>

</head>

<body>
	<form name="searchForm">
		<table id="searchTable" align="left" valign="middle" cellspacing="0" cellpadding="0" border="0">
			<tr nowrap  valign="middle">
				<td  id="searchTD">
					<label id="searchLabel" for="searchWord" accesskey="s">
					&nbsp;Search:
					</label>
				</td>
				<td>
					<input type="text" id="searchWord" name="searchWord" value='' size="24" maxlength="256" 
					       alt="&#42; &#61; any string&#44; &#63; &#61; any character&#44; &#34;&#34; &#61; phrase&#44; AND&#44; OR&#44; NOT &#61; boolean operators " 
					       title="&#42; &#61; any string&#44; &#63; &#61; any character&#44; &#34;&#34; &#61; phrase&#44; AND&#44; OR&#44; NOT &#61; boolean operators ">
				</td>
				<td >
					<input type="button" onclick="this.blur();" value="Go" id="go" alt="Go" title="Go">
					<input type="hidden" name="maxHits" value="500" >
				</td>
				<td nowrap>
					<a id="scopeLabel" href="javascript:openAdvanced();" title="Select Scope" alt="Select Scope">Scope:</a>
				</td>
				<td nowrap>
					<input type="hidden" name="workingSet" value='All topics'>
					<div id="scope" >All topics</div>
				</td>
			</tr>

		</table>
	</form>		

</body>
</html>
