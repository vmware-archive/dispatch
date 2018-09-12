if (/iP(?:od|hone)/.test(navigator.userAgent)) {
  // deletable at least, i.e.
  //  if (window.gestureStart && dontWantIt) {
  //    document.removeEventListener("gesturestart", gestureStart);
  //    delete window.gestureStart;
  //  }
  this.gestureStart = (function(metas, i){
    function gestureStart() {
      // address .length once per loop
      for (i = metas.length; i--;) {
        if (metas[i].name == "viewport") {
          metas[i].content = "width=device-width, minimum-scale=0.25, maximum-scale=1.6";
        }
      }
    }
    // address .length once per loop
    for (i = metas.length; i--;) {
      if (metas[i].name == "viewport") {
        metas[i].content = "width=device-width, minimum-scale=1.0, maximum-scale=1.0";
      }
    }
    // I might argue it should be on capture phase here
    document.addEventListener("gesturestart", gestureStart);
    // to be able to address it out there
    return gestureStart;
  }(document.getElementsByTagName('meta')));
}
