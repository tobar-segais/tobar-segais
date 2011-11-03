<!doctype html>
<%--
  ~ Copyright 2011 Stephen Connolly
  ~
  ~ Licensed under the Apache License, Version 2.0 (the "License");
  ~ you may not use this file except in compliance with the License.
  ~ You may obtain a copy of the License at
  ~
  ~     http://www.apache.org/licenses/LICENSE-2.0
  ~
  ~ Unless required by applicable law or agreed to in writing, software
  ~ distributed under the License is distributed on an "AS IS" BASIS,
  ~ WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  ~ See the License for the specific language governing permissions and
  ~ limitations under the License.
  --%>
<html lang="en">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=ISO-8859-1">
<title>Navigation Views</title>
<style type="text/css">

/* need this one for Mozilla */
HTML {
	width:100%;
	height:100%;
	margin:0px;
	padding:0px;
	border:0px;
 }

BODY {
	margin:0px;
	padding:0px;
	/* Mozilla does not like width:100%, so we set height only */
	height:100%;
	position : relative;  // Needed for Safari
}

IFRAME {
	width:100%;
	height:100%;
	position : absolute;  // Needed for Safari
	top : 0px;
}

.hidden {
	visibility:hidden;
	width:0;
	height:0;
}

.visible {
	visibility:visible;
	width:100%;
	height:100%;
}

</style>
</head>
<body dir="ltr" tabIndex="-1" onunload="closeConfirmShowAllDialog()">
 	<iframe frameborder="0"
 		    class="visible"
 		    name="toc"
 		    title="Layout frame: toc"
 		    id="toc"
 		    scrolling="yes"
 		    src='toc.jsp'>
 	</iframe>

 	<iframe frameborder="0"
 		    class="hidden"
 		    name="index"
 		    title="Layout frame: index"
 		    id="index"
 		    scrolling="no"
 		    src='view.jsp?view=index'>
 	</iframe>

 	<iframe frameborder="0"
 		    class="hidden"
 		    name="search"
 		    title="Layout frame: search"
 		    id="search"
 		    scrolling="no"
 		    src='view.jsp?view=search'>
 	</iframe>


</body>
</html>

