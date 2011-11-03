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
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
<title>Tabs</title>
<style type="text/css">

body {
	margin:0;
	padding:0;
	height:100%;
    height:23px;
}

a {
    cursor : default
}

/* tabs at the bottom */
.tab {
	font-size:5px;
	margin:0;
	padding:0;
	border-top:1px solid ThreeDShadow;
	border-bottom:1px solid ButtonFace;
	cursor:default;
	background:ButtonFace;
}

.pressed {
	font-size:5px;
	margin:0;
	padding:0;
	cursor:default;
	
	border-top:0 solid ButtonFace;
	border-bottom:1px solid ThreeDShadow;
}

.separator {
	height:100%;
	background-color:ThreeDShadow;
	border-bottom:1px solid ButtonFace;
}

.separator_pressed {
	height:100%;
	background-color:ThreeDShadow;
	border-top:0 solid ButtonFace;
	border-bottom:1px solid ButtonFace;
}

a {
	text-decoration:none;
	vertical-align:middle;
	height:16px;
	width:16px;

	display:block;

}

img {
	border:0;
	margin:0;
	padding:0;
	height:16px;
	width:16px;
}

</style>
 
</head>
<body>
  <table cellspacing="0" cellpadding="0" border="0" width="100%" height="100%" valign="middle">
   <tr>
	<td  title="Contents"
	     align="center"  
	     valign="middle"
	     class="tab" 
	     id="toc" 
	     onclick="parent.showView('toc')" 
	     onmouseover="window.status='Contents';return true;" 
	     onmouseout="window.status='';">
	     <a  href='javascript:parent.showView("toc");' 
	         onclick='this.blur();return false;' 
	         onmouseover="window.status='Contents';return true;" 
	         onmouseout="window.status='';"
	         id="linktoc"
	         accesskey="C">
	         <img alt="Contents" 
	              title="Contents" 
	              src="images/contents_view.png"
	              id="imgtoc"
	              height="16"
	         >
	     </a>
	</td>

	<td width="1px" class="separator"><div style="width:1px;height:1px;display:block;"></div></td>

	<td  title="Index" 
	     align="center"  
	     valign="middle"
	     class="tab" 
	     id="index" 
	     onclick="parent.showView('index')" 
	     onmouseover="window.status='Index';return true;" 
	     onmouseout="window.status='';">
	     <a  href='javascript:parent.showView("index");' 
	         onclick='this.blur();return false;' 
	         onmouseover="window.status='Index';return true;" 
	         onmouseout="window.status='';"
	         id="linkindex"
	         ACCESSKEY="I">
	         <img alt="Index" 
	              title="Index" 
	              src="images/index_view.png"
	              id="imgindex"
	              height="16"
	         >
	     </a>
	</td>

	<td width="1px" class="separator"><div style="width:1px;height:1px;display:block;"></div></td>

	<td  title="Search Results" 
	     align="center"  
	     valign="middle"
	     class="tab" 
	     id="search" 
	     onclick="parent.showView('search')" 
	     onmouseover="window.status='Search\u0020Results';return true;" 
	     onmouseout="window.status='';">
	     <a  href='javascript:parent.showView("search");' 
	         onclick='this.blur();return false;' 
	         onmouseover="window.status='Search\u0020Results';return true;" 
	         onmouseout="window.status='';"
	         id="linksearch"
	         ACCESSKEY="R">
	         <img alt="Search Results" 
	              title="Search Results" 
	              src="images/search_view.png"
	              id="imgsearch"
	              height="16"
	         >
	     </a>
	</td>

   </tr>
   </table>

</body>
</html>

