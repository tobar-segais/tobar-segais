/*
 * Copyright 2012 Stephen Connolly
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

var TobairSegais = {
    toRawUri:function (uri) {
        var i = uri.indexOf('#');
        if (i != -1) {
            uri = uri.substring(0, i);
        }
        i = uri.indexOf('?');
        if (i != -1) {
            uri = uri.substring(0, i);
        }
        return uri + "?raw";
    },
    clickSupport:function () {
        if ($(this).attr("ts-immediate") == "true") return true;
        return TobairSegais.loadContent($(this).attr('href'));
    },
    addClickSupport:function (id) {
        $(id + ' a').each(function () {
            $(this).click(TobairSegais.clickSupport);
        });
    },
    loadContent:function (url) {
        if (/^https?:\/\//.test(url)) {
            return true;
        } else {
            $('#content').load(TobairSegais.toRawUri(url), function () {
                TobairSegais.addClickSupport("#content");
                history.pushState({url:url}, "", url);
                var i = url.indexOf('#');
                if (i != -1) {
                    window.location.href = url.substring(i);
                }
                document.title = $("#contents-nav a[href='"+(i == -1 ? url : url.substring(0,i))+"']").text();
            });
            return false;
        }
    },
    windowSizer:function(){
      var css = {'height':'100%','overflow':'auto','margin':0,'padding':0,'position':'relative','border':'none', 'border-redius':0};
      
      var $win = $(window); 
      var $top = $('.navbar-fixed-top').first().css(css);
      var $bottom = $('.navbar-fixed-bottom').first().css(css);
      var $body = $('.row-fluid').first().css(css);
      var $sideBar = $('.sidebar-nav').first().parent().css(css);
      var $content = $('#content').css('padding','20px 5% 20px 0').parent().css(css).css('float','right');
      var $sbContent = $('#sidebar-content').css(css);
      
      $('body').css(css);
      
      $body.height($win.height() - ($top.height() + $bottom.height()) ).parent().css({'padding':0});
      $sbContent.height($win.height() - $sbContent.offset().top).parent().css(css);
      $('.sidebar-nav').first().css({'border-right':'1px solid #ccc', 'border-radius':0});
      
      if($win.width() < 768){
        $sideBar.add($sbContent).css({'height':'auto','overflow':'auto'});
        $('#content').css('padding','20px');
      }
    }
};
window.onpopstate = function (event) {
    if (event != null && event.state != null) {
        $('#content').load(TobairSegais.toRawUri(event.state.url), function () {
            TobairSegais.addClickSupport("#content");
        });
    }
};
$(function () {
    TobairSegais.addClickSupport("#content");
    TobairSegais.addClickSupport("#sidebar-content");
    TobairSegais.windowSizer();
    $(window).resize(TobairSegais.windowSizer);
});
