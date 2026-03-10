use adw::subclass::prelude::*;
use gtk::glib;
use gtk::prelude::*;

mod imp {
    use super::*;

    #[derive(Debug, Default, gtk::CompositeTemplate)]
    #[template(file = "../window.ui")]
    pub struct TrucoWindow {
        #[template_child]
        pub main_overlay: TemplateChild<gtk::Overlay>,
        #[template_child]
        pub hud_box: TemplateChild<gtk::Box>,
        #[template_child]
        pub opponent_box: TemplateChild<gtk::Box>,
        #[template_child]
        pub center_box: TemplateChild<gtk::Box>,
        #[template_child]
        pub bottom_box: TemplateChild<gtk::Box>,
    }

    #[glib::object_subclass]
    impl ObjectSubclass for TrucoWindow {
        const NAME: &'static str = "TrucoWindow";
        type Type = super::TrucoWindow;
        type ParentType = adw::ApplicationWindow;

        fn class_init(klass: &mut Self::Class) {
            klass.bind_template();
        }

        fn instance_init(obj: &glib::subclass::InitializingObject<Self>) {
            obj.init_template();
        }
    }

    impl ObjectImpl for TrucoWindow {}
    impl WidgetImpl for TrucoWindow {}
    impl WindowImpl for TrucoWindow {}
    impl ApplicationWindowImpl for TrucoWindow {}
    impl adw::subclass::prelude::AdwApplicationWindowImpl for TrucoWindow {}
}

glib::wrapper! {
    pub struct TrucoWindow(ObjectSubclass<imp::TrucoWindow>)
        @extends gtk::Widget, gtk::Window, gtk::ApplicationWindow, adw::ApplicationWindow,
        @implements gio::ActionMap, gio::ActionGroup, gtk::Root;
}

impl TrucoWindow {
    pub fn new(app: &adw::Application) -> Self {
        glib::Object::builder().property("application", app).build()
    }

    pub fn main_overlay(&self) -> gtk::Overlay {
        self.imp().main_overlay.get()
    }
    
    pub fn hud_box(&self) -> gtk::Box {
        self.imp().hud_box.get()
    }

    pub fn opponent_box(&self) -> gtk::Box {
        self.imp().opponent_box.get()
    }

    pub fn center_box(&self) -> gtk::Box {
        self.imp().center_box.get()
    }

    pub fn bottom_box(&self) -> gtk::Box {
        self.imp().bottom_box.get()
    }
}
