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
                } else {
                    window.location.href = "#";
                }
                document.title = $("#contents-nav a[href='"+(i == -1 ? url : url.substring(0,i))+"']").text();
            });
            return false;
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
});
